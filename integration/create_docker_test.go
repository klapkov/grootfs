package integration_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	yaml "gopkg.in/yaml.v2"

	"code.cloudfoundry.org/grootfs/commands/config"
	"code.cloudfoundry.org/grootfs/groot"
	runnerpkg "code.cloudfoundry.org/grootfs/integration/runner"
	"code.cloudfoundry.org/grootfs/store"
	"code.cloudfoundry.org/grootfs/testhelpers"
	"code.cloudfoundry.org/lager"

	"github.com/alecthomas/units"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Create with remote DOCKER images", func() {
	var (
		randomImageID string
		baseImageURL  string
		runner        runnerpkg.Runner
	)

	BeforeEach(func() {
		initSpec := runnerpkg.InitSpec{UIDMappings: []groot.IDMappingSpec{
			{HostID: GrootUID, NamespaceID: 0, Size: 1},
			{HostID: 100000, NamespaceID: 1, Size: 65000},
		},
			GIDMappings: []groot.IDMappingSpec{
				{HostID: GrootGID, NamespaceID: 0, Size: 1},
				{HostID: 100000, NamespaceID: 1, Size: 65000},
			},
		}

		randomImageID = testhelpers.NewRandomID()
		Expect(Runner.RunningAsUser(0, 0).InitStore(initSpec)).To(Succeed())
		runner = Runner.SkipInitStore()
	})

	Context("when using the default registry", func() {
		BeforeEach(func() {
			baseImageURL = "docker:///cfgarden/empty:v0.1.0"
		})

		It("creates a root filesystem based on the image provided", func() {
			containerSpec, err := runner.Create(groot.CreateSpec{
				BaseImage: "docker:///cfgarden/three-layers",
				ID:        randomImageID,
				Mount:     mountByDefault(),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.EnsureMounted(containerSpec)).To(Succeed())

			Expect(path.Join(containerSpec.Root.Path, "layer-3-file")).To(BeARegularFile())
			Expect(path.Join(containerSpec.Root.Path, "layer-2-file")).To(BeARegularFile())
			Expect(path.Join(containerSpec.Root.Path, "layer-1-file")).To(BeARegularFile())
		})

		It("gives any user permission to be inside the container", func() {
			containerSpec, err := runner.Create(groot.CreateSpec{
				BaseImage: "docker:///cfgarden/garden-busybox",
				ID:        randomImageID,
				Mount:     mountByDefault(),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.EnsureMounted(containerSpec)).To(Succeed())

			cmd := exec.Command(NamespacerBin, containerSpec.Root.Path, strconv.Itoa(GrootUID+100), "/bin/ls", "/")
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
			}
			sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, 5*time.Second).Should(gexec.Exit(0))
		})

		It("outputs a json with the correct `Root.Path` key", func() {
			containerSpec, err := runner.Create(groot.CreateSpec{
				BaseImage: "docker:///cfgarden/garden-busybox",
				ID:        randomImageID,
				Mount:     mountByDefault(),
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(containerSpec.Root.Path).To(Equal(filepath.Join(StorePath, store.ImageDirName, randomImageID, "rootfs")))
		})

		It("outputs a json with the correct `Process.Env` key", func() {
			containerSpec, err := runner.Create(groot.CreateSpec{
				BaseImage: "docker:///cfgarden/with-volume",
				ID:        randomImageID,
				Mount:     mountByDefault(),
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(containerSpec.Process.Env).To(ContainElement("PATH=/usr/local/bin:/usr/bin:/bin:/from-dockerfile"))
			Expect(containerSpec.Process.Env).To(ContainElement("TEST=second-test-from-dockerfile:test-from-dockerfile"))
		})

		Context("when the image is bigger than available memory", func() {
			var tcpDumpSess *gexec.Session

			BeforeEach(func() {
				var err error
				tcpDumpSess, err = gexec.Start(exec.Command("tcpdump", "-U", "-w", "/tmp/packets", "tcp"), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				tcpDumpSess.Kill()
			})

			It("doesn't fail", func() {
				sess, err := runner.StartCreate(groot.CreateSpec{
					BaseImage: "docker:///ubuntu:trusty",
					ID:        "some-id",
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())

				go func() {
					defer GinkgoRecover()

					statsPath := path.Join("/proc", strconv.Itoa(sess.Command.Process.Pid), "status")
					runs := 0
					for {
						stats, err := ioutil.ReadFile(statsPath)
						if err != nil {
							Expect(runs).To(BeNumerically(">", 1))
							break
						}

						var statsMap map[string]string
						Expect(yaml.Unmarshal(stats, &statsMap)).To(Succeed())

						virtualMemoryHighWaterMark := strings.Replace(strings.ToUpper(statsMap["VmHWM"]), " ", "", -1)
						if virtualMemoryHighWaterMark != "" {
							n, err := units.ParseBase2Bytes(virtualMemoryHighWaterMark)
							Expect(err).NotTo(HaveOccurred())
							// Biggest ubuntu:trusty layer is 65694192 bytes
							Expect(n).To(BeNumerically("<", 50*1024*1024))
						}

						time.Sleep(200 * time.Millisecond)
						runs++
					}
				}()

				deadline := time.Now().Add(60 * time.Second)
				for {
					if sess.ExitCode() != -1 {
						break
					}
					if time.Now().After(deadline) {
						fmt.Println(">>>> printing debug info")

						fmt.Println(">>>> netstat:")
						netstatCmd := exec.Command("bash", "-c", "netstat -tanp | grep "+strconv.Itoa(sess.Command.Process.Pid))
						netstatCmd.Stdout = GinkgoWriter
						netstatCmd.Stderr = GinkgoWriter
						Expect(netstatCmd.Run()).To(Succeed())

						fmt.Println(">>>> tcpdump:")
						cmd := exec.Command("tcpdump", "-r", "/tmp/packets", "(src registry-1.docker.io or dst registry-1.docker.io) and tcp[tcpflags] & (tcp-ack) != 0")
						cmd.Stdout = GinkgoWriter
						cmd.Stderr = GinkgoWriter
						Expect(cmd.Run()).To(Succeed())
						fmt.Println(">>>> if this build recently failed, please hijack the container and download `/tmp/packets`, which contains more tcpdump info")

						fmt.Println(">>>> goroutine stack:")
						sess.Signal(syscall.SIGQUIT)
						Fail("timeout exeeded")
					}
					time.Sleep(100 * time.Millisecond)
				}
				Expect(sess.ExitCode()).To(Equal(0))
			})
		})

		Context("when the image has volumes", func() {
			BeforeEach(func() {
				baseImageURL = "docker:///cfgarden/with-volume"
			})

			It("lists the volumes as mounts in the returned spec", func() {
				containerSpec, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())

				volumeHash := sha256.Sum256([]byte("/foo"))
				mountSourceName := "vol-" + hex.EncodeToString(volumeHash[:32])
				Expect(containerSpec.Mounts).To(ContainElement(specs.Mount{
					Destination: "/foo",
					Source:      filepath.Join(StorePath, store.ImageDirName, randomImageID, mountSourceName),
					Type:        "bind",
					Options:     []string{"bind"},
				}))
			})

			It("create the bind mount source for the volumes", func() {
				_, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())

				volumeHash := sha256.Sum256([]byte("/foo"))
				mountSourceName := "vol-" + hex.EncodeToString(volumeHash[:32])

				Expect(filepath.Join(StorePath, store.ImageDirName, randomImageID, mountSourceName)).To(BeADirectory())
			})
		})

		Context("when the image has links that overwrites existing files", func() {
			BeforeEach(func() {
				baseImageURL = "docker:///cfgarden/overwrite-link"
			})

			It("creates the link with success", func() {
				containerSpec, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.EnsureMounted(containerSpec)).To(Succeed())

				symlinkFilePath := filepath.Join(containerSpec.Root.Path, "tmp/symlink")
				stat, err := os.Lstat(symlinkFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(stat.Mode() & os.ModeSymlink).ToNot(BeZero())
				linkTargetPath, err := os.Readlink(symlinkFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(linkTargetPath).To(Equal("/etc/link-source"))
			})
		})

		Context("when the image has opaque white outs", func() {
			BeforeEach(func() {
				baseImageURL = "docker:///cfgarden/opq-whiteout-busybox"
			})

			It("empties the folder contents but keeps the dir", func() {
				containerSpec, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.EnsureMounted(containerSpec)).To(Succeed())

				whiteoutedDir := path.Join(containerSpec.Root.Path, "var")
				Expect(whiteoutedDir).To(BeADirectory())
				contents, err := ioutil.ReadDir(whiteoutedDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(BeEmpty())
			})
		})

		Context("when the image has whiteouts", func() {
			BeforeEach(func() {
				baseImageURL = "docker:///cfgarden/with-whiteouts"
			})

			It("removes the whiteout file", func() {
				containerSpec, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.EnsureMounted(containerSpec)).To(Succeed())

				Expect(path.Join(containerSpec.Root.Path, "file-to-be-deleted")).ToNot(BeAnExistingFile())
				Expect(path.Join(containerSpec.Root.Path, "folder")).ToNot(BeAnExistingFile())
				Expect(path.Join(containerSpec.Root.Path, "existing-file")).To(BeAnExistingFile())
			})
		})

		Context("when the image has files with the setuid on", func() {
			BeforeEach(func() {
				baseImageURL = "docker:///cfgarden/garden-busybox"
			})

			It("correctly applies the user bit", func() {
				containerSpec, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.EnsureMounted(containerSpec)).To(Succeed())

				setuidFilePath := path.Join(containerSpec.Root.Path, "bin", "busybox")
				stat, err := os.Stat(setuidFilePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(stat.Mode() & os.ModeSetuid).To(Equal(os.ModeSetuid))
			})
		})

		Describe("clean up on create", func() {
			JustBeforeEach(func() {
				_, err := runner.Create(groot.CreateSpec{
					ID:        "my-busybox",
					BaseImage: "docker:///cfgarden/garden-busybox",
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(runner.Delete("my-busybox")).To(Succeed())
			})

			AfterEach(func() {
				Expect(runner.Delete(randomImageID)).To(Succeed())
			})

			It("cleans up unused layers before create but not the one about to be created", func() {
				createSpec := groot.CreateSpec{
					ID:        "my-empty",
					BaseImage: "docker:///cfgarden/empty:v0.1.1",
					Mount:     mountByDefault(),
				}
				_, err := runner.Create(createSpec)
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.Delete("my-empty")).To(Succeed())

				layerPath := filepath.Join(StorePath, store.VolumesDirName, testhelpers.EmptyBaseImageV011.Layers[0].ChainID)
				stat, err := os.Stat(layerPath)
				Expect(err).NotTo(HaveOccurred())
				preLayerTimestamp := stat.ModTime()

				preContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
				Expect(err).NotTo(HaveOccurred())
				Expect(preContents).To(HaveLen(3))

				runner := runner.WithClean()
				_, err = runner.Create(groot.CreateSpec{
					ID:        randomImageID,
					BaseImage: "docker:///cfgarden/empty:v0.1.1",
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())

				afterContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
				Expect(err).NotTo(HaveOccurred())
				Expect(afterContents).To(HaveLen(2))

				for _, layer := range testhelpers.EmptyBaseImageV011.Layers {
					Expect(filepath.Join(StorePath, store.VolumesDirName, layer.ChainID)).To(BeADirectory())
				}

				stat, err = os.Stat(layerPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(stat.ModTime()).To(Equal(preLayerTimestamp))
			})

			Context("when no-clean flag is set", func() {
				It("does not clean up unused layers", func() {
					preContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
					Expect(err).NotTo(HaveOccurred())
					Expect(preContents).To(HaveLen(1))

					_, err = runner.WithNoClean().Create(groot.CreateSpec{
						ID:        randomImageID,
						BaseImage: "docker:///cfgarden/empty:v0.1.1",
						Mount:     mountByDefault(),
					})
					Expect(err).NotTo(HaveOccurred())

					afterContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
					Expect(err).NotTo(HaveOccurred())
					Expect(afterContents).To(HaveLen(3))
				})
			})
		})

		Context("when the image is private", func() {
			BeforeEach(func() {
				baseImageURL = "docker:///cfgarden/private"
			})

			Context("when the credentials are correct", func() {
				It("succeeds", func() {
					runner := runner.WithCredentials(RegistryUsername, RegistryPassword)
					containerSpec, err := runner.Create(groot.CreateSpec{
						BaseImage: baseImageURL,
						ID:        randomImageID,
						Mount:     mountByDefault(),
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(runner.EnsureMounted(containerSpec)).To(Succeed())
				})
			})

			Context("when the credentials are wrong", func() {
				// We need a fake registry here because Dockerhub was rate limiting on multiple bad credential auth attempts
				var fakeRegistry *testhelpers.FakeRegistry

				BeforeEach(func() {
					dockerHubUrl, err := url.Parse("https://registry-1.docker.io")
					Expect(err).NotTo(HaveOccurred())
					fakeRegistry = testhelpers.NewFakeRegistry(dockerHubUrl)
					fakeRegistry.Start()
					fakeRegistry.ForceTokenAuthError()
					baseImageURL = fmt.Sprintf("docker://%s/doesnt-matter-because-fake-registry", fakeRegistry.Addr())
				})

				AfterEach(func() {
					fakeRegistry.Stop()
				})

				It("fails", func() {
					runner := runner.WithCredentials("someuser", "invalid-password").WithInsecureRegistry(fakeRegistry.Addr())
					_, err := runner.Create(groot.CreateSpec{
						BaseImage: baseImageURL,
						ID:        randomImageID,
						Mount:     mountByDefault(),
					})
					Expect(err).To(MatchError(ContainSubstring("authorization failed: username and password are invalid")))
				})
			})

			It("does not log the credentials OR their references", func() {
				buffer := gbytes.NewBuffer()
				runner := runner.WithCredentials(RegistryUsername, RegistryPassword).WithStderr(buffer).WithLogLevel(lager.DEBUG)
				_, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).ToNot(HaveOccurred())
				Eventually(buffer).ShouldNot(gbytes.Say("\"RegistryUsername\":\"\",\"RegistryPassword\":\"\""))
				Eventually(buffer).ShouldNot(gbytes.Say(RegistryPassword))
			})
		})

		Context("when image size exceeds disk quota", func() {
			BeforeEach(func() {
				baseImageURL = "docker:///cfgarden/empty:v0.1.1"
			})

			Context("when the image is not accounted for in the quota", func() {
				It("succeeds", func() {
					containerSpec, err := runner.Create(groot.CreateSpec{
						BaseImage: baseImageURL,
						ID:        randomImageID,
						Mount:     mountByDefault(),
						ExcludeBaseImageFromQuota: true,
						DiskLimit:                 10,
					})
					Expect(runner.EnsureMounted(containerSpec)).To(Succeed())
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when the image is accounted for in the quota", func() {
				It("returns an error", func() {
					_, err := runner.Create(groot.CreateSpec{
						BaseImage: baseImageURL,
						ID:        randomImageID,
						Mount:     mountByDefault(),
						DiskLimit: 10,
					})
					Expect(err).To(MatchError(ContainSubstring("layers exceed disk quota")))
				})
			})
		})

		Describe("Unpacked layer caching", func() {
			It("caches the unpacked image as a volume", func() {
				_, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).ToNot(HaveOccurred())

				layerSnapshotPath := filepath.Join(StorePath, "volumes", "3355e23c079e9b35e4b48075147a7e7e1850b99e089af9a63eed3de235af98ca")
				Expect(ioutil.WriteFile(layerSnapshotPath+"/injected-file", []byte{}, 0666)).To(Succeed())

				containerSpec, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        testhelpers.NewRandomID(),
					Mount:     mountByDefault(),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(runner.EnsureMounted(containerSpec)).To(Succeed())
				Expect(path.Join(containerSpec.Root.Path, "hello")).To(BeARegularFile())
				Expect(path.Join(containerSpec.Root.Path, "injected-file")).To(BeARegularFile())
			})

			Describe("when one of the layers is corrupted", func() {
				var (
					fakeRegistry  *testhelpers.FakeRegistry
					corruptedBlob string
					volumesDir    string
				)

				AfterEach(func() {
					fakeRegistry.Stop()
				})

				Context("when the image has a version 2 manifest schema", func() {
					BeforeEach(func() {
						volumesDir = filepath.Join(StorePath, store.VolumesDirName)
						dockerHubUrl, err := url.Parse("https://registry-1.docker.io")
						Expect(err).NotTo(HaveOccurred())
						fakeRegistry = testhelpers.NewFakeRegistry(dockerHubUrl)
						corruptedBlob = testhelpers.EmptyBaseImageV011.Layers[1].BlobID
						fakeRegistry.WhenGettingBlob(corruptedBlob, 0, func(w http.ResponseWriter, r *http.Request) {
							_, err := w.Write([]byte("bad-blob"))
							Expect(err).NotTo(HaveOccurred())
						})
						fakeRegistry.Start()
						baseImageURL = fmt.Sprintf("docker://%s/cfgarden/empty:v0.1.1", fakeRegistry.Addr())
					})

					It("fails and cleans up the corrupted volumes", func() {
						runner := runner.WithInsecureRegistry(fakeRegistry.Addr())

						_, err := runner.Create(groot.CreateSpec{
							BaseImage: baseImageURL,
							ID:        randomImageID,
							Mount:     mountByDefault(),
						})

						Expect(err).To(MatchError(ContainSubstring("layer is corrupted")))

						volumes, _ := ioutil.ReadDir(volumesDir)
						Expect(len(volumes)).To(Equal(len(testhelpers.EmptyBaseImageV011.Layers) - 1))

						Expect(filepath.Join(volumesDir, testhelpers.EmptyBaseImageV011.Layers[0].ChainID)).To(BeADirectory())
						Expect(filepath.Join(volumesDir, testhelpers.EmptyBaseImageV011.Layers[1].ChainID)).ToNot(BeADirectory())
					})
				})

				Context("when the image has a version 1 manifest schema", func() {
					BeforeEach(func() {
						dockerHubUrl, err := url.Parse("https://registry-1.docker.io")
						Expect(err).NotTo(HaveOccurred())
						fakeRegistry = testhelpers.NewFakeRegistry(dockerHubUrl)
						corruptedBlob = testhelpers.SchemaV1EmptyBaseImage.Layers[2].BlobID
						fakeRegistry.WhenGettingBlob(corruptedBlob, 0, func(w http.ResponseWriter, r *http.Request) {
							_, err := w.Write([]byte("bad-blob"))
							Expect(err).NotTo(HaveOccurred())
						})
						fakeRegistry.Start()
						baseImageURL = fmt.Sprintf("docker://%s/cfgarden/empty:schemaV1", fakeRegistry.Addr())
					})

					It("fails and never creates any volumes", func() {
						runner := runner.WithInsecureRegistry(fakeRegistry.Addr())

						_, err := runner.Create(groot.CreateSpec{
							BaseImage: baseImageURL,
							ID:        randomImageID,
							Mount:     mountByDefault(),
						})

						Expect(err).To(MatchError(ContainSubstring("converting V1 schema failed")))

						volumes, _ := ioutil.ReadDir(volumesDir)
						Expect(len(volumes)).To(Equal(0))
					})
				})
			})
		})

		Context("when the image has a version 1 manifest schema", func() {
			BeforeEach(func() {
				baseImageURL = "docker:///cfgarden/empty:schemaV1"
			})

			It("creates a root filesystem based on the image provided", func() {
				containerSpec, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(runner.EnsureMounted(containerSpec)).To(Succeed())
				Expect(path.Join(containerSpec.Root.Path, "allo")).To(BeAnExistingFile())
				Expect(path.Join(containerSpec.Root.Path, "hello")).To(BeAnExistingFile())

				volumesDir := filepath.Join(StorePath, store.VolumesDirName)
				Expect(filepath.Join(volumesDir, testhelpers.SchemaV1EmptyBaseImage.Layers[0].ChainID)).To(BeADirectory())
				Expect(filepath.Join(volumesDir, testhelpers.SchemaV1EmptyBaseImage.Layers[1].ChainID)).To(BeADirectory())
				Expect(filepath.Join(volumesDir, testhelpers.SchemaV1EmptyBaseImage.Layers[2].ChainID)).To(BeADirectory())
			})
		})
	})

	Context("when a private registry is used", func() {
		var (
			fakeRegistry *testhelpers.FakeRegistry
		)

		BeforeEach(func() {
			dockerHubUrl, err := url.Parse("https://registry-1.docker.io")
			Expect(err).NotTo(HaveOccurred())
			fakeRegistry = testhelpers.NewFakeRegistry(dockerHubUrl)
			fakeRegistry.Start()

			baseImageURL = fmt.Sprintf("docker://%s/cfgarden/empty:v0.1.1", fakeRegistry.Addr())
		})

		AfterEach(func() {
			fakeRegistry.Stop()
		})

		It("fails to fetch the image", func() {
			_, err := runner.Create(groot.CreateSpec{
				BaseImage: baseImageURL,
				ID:        randomImageID,
				Mount:     mountByDefault(),
			})
			Expect(err).To(MatchError("This registry is insecure. To pull images from this registry, please use the --insecure-registry option."))
		})

		Context("when it's provided as a valid insecure registry", func() {
			It("should create a root filesystem based on the image provided by the private registry", func() {
				runner := runner.WithInsecureRegistry(fakeRegistry.Addr())
				containerSpec, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.EnsureMounted(containerSpec)).To(Succeed())

				Expect(path.Join(containerSpec.Root.Path, "hello")).To(BeARegularFile())
				Expect(fakeRegistry.RequestedBlobs()).To(HaveLen(3))
			})
		})

		Context("when --config global flag is given", func() {
			Describe("with an insecure registries list", func() {
				var (
					configDir      string
					configFilePath string
				)

				BeforeEach(func() {
					var err error
					configDir, err = ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())
					Expect(os.Chmod(configDir, 0755)).To(Succeed())

					cfg := config.Config{
						Create: config.Create{
							InsecureRegistries: []string{fakeRegistry.Addr()},
						},
					}

					configYaml, err := yaml.Marshal(cfg)
					Expect(err).NotTo(HaveOccurred())
					configFilePath = path.Join(configDir, "config.yaml")

					Expect(ioutil.WriteFile(configFilePath, configYaml, 0755)).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(configDir)).To(Succeed())
				})

				It("creates a root filesystem based on the image provided by the private registry", func() {
					runner := runner.WithConfig(configFilePath)
					containerSpec, err := runner.Create(groot.CreateSpec{
						BaseImage: baseImageURL,
						ID:        randomImageID,
						Mount:     mountByDefault(),
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(runner.EnsureMounted(containerSpec)).To(Succeed())

					Expect(path.Join(containerSpec.Root.Path, "hello")).To(BeARegularFile())
					Expect(fakeRegistry.RequestedBlobs()).To(HaveLen(3))
				})
			})

			Context("when config path is invalid", func() {
				It("returns a useful error", func() {
					runner := runner.WithConfig("invalid-config-path")
					_, err := runner.Create(groot.CreateSpec{
						BaseImage: baseImageURL,
						ID:        randomImageID,
						Mount:     mountByDefault(),
					})
					Expect(err).To(MatchError(ContainSubstring("invalid config path")))
				})
			})
		})
	})

	Context("when the image does not exist", func() {
		It("returns a useful error", func() {
			_, err := runner.Create(groot.CreateSpec{
				BaseImage: "docker:///cfgaren/sorry-not-here",
				ID:        randomImageID,
				Mount:     mountByDefault(),
			})
			Expect(err).To(MatchError(ContainSubstring("docker:///cfgaren/sorry-not-here does not exist or you do not have permissions to see it.")))
		})
	})

	Context("when the image has files that are not writable to their owner", func() {
		BeforeEach(func() {
			baseImageURL = "docker:///cfgarden/non-writable-file"
		})

		Context("id mappings", func() {
			It("works", func() {
				containerSpec, err := runner.Create(groot.CreateSpec{
					BaseImage: baseImageURL,
					ID:        randomImageID,
					Mount:     mountByDefault(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.EnsureMounted(containerSpec)).To(Succeed())

				Expect(path.Join(containerSpec.Root.Path, "test", "hello")).To(BeARegularFile())
			})
		})
	})

	Context("when the image has folders that are not writable to their owner", func() {
		BeforeEach(func() {
			baseImageURL = "docker:///cfgarden/non-writable-folder"
		})

		It("works", func() {
			containerSpec, err := runner.Create(groot.CreateSpec{
				BaseImage: baseImageURL,
				ID:        randomImageID,
				Mount:     mountByDefault(),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.EnsureMounted(containerSpec)).To(Succeed())
			Expect(path.Join(containerSpec.Root.Path, "test", "hello")).To(BeARegularFile())
		})
	})

	Context("when --skip-layer-validation flag is passed", func() {
		var (
			fakeRegistry  *testhelpers.FakeRegistry
			corruptedBlob string
		)

		AfterEach(func() {
			fakeRegistry.Stop()
		})

		BeforeEach(func() {
			dockerHubUrl, err := url.Parse("https://registry-1.docker.io")
			Expect(err).NotTo(HaveOccurred())
			fakeRegistry = testhelpers.NewFakeRegistry(dockerHubUrl)
			corruptedBlob = testhelpers.EmptyBaseImageV011.Layers[1].BlobID
			fakeRegistry.WhenGettingBlob(corruptedBlob, 0, func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write([]byte("bad-blob"))
				Expect(err).NotTo(HaveOccurred())
			})
			fakeRegistry.Start()
			baseImageURL = fmt.Sprintf("docker://%s/cfgarden/empty:v0.1.1", fakeRegistry.Addr())
		})

		It("has no effect", func() {
			r := runner.WithInsecureRegistry(fakeRegistry.Addr())
			_, err := r.SkipLayerCheckSumValidation().Create(groot.CreateSpec{
				BaseImage: baseImageURL,
				ID:        randomImageID,
				Mount:     true,
			})

			Expect(err).To(MatchError(ContainSubstring("layer is corrupted")))
		})
	})
})
