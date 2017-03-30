package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/grootfs/commands/config"
	"code.cloudfoundry.org/grootfs/groot"
	"code.cloudfoundry.org/grootfs/integration"
	"code.cloudfoundry.org/grootfs/store"
	"code.cloudfoundry.org/grootfs/store/filesystems/overlayxfs"
	"code.cloudfoundry.org/grootfs/testhelpers"
	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const (
	tenMegabytes = int64(10485760)
)

var _ = Describe("Create", func() {
	var (
		baseImagePath   string
		sourceImagePath string
		rootUID         int
		rootGID         int
	)

	BeforeEach(func() {
		rootUID = 0
		rootGID = 0

		var err error
		sourceImagePath, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chown(sourceImagePath, rootUID, rootGID)).To(Succeed())
		Expect(os.Chmod(sourceImagePath, 0755)).To(Succeed())

		grootFilePath := path.Join(sourceImagePath, "foo")
		Expect(ioutil.WriteFile(grootFilePath, []byte("hello-world"), 0644)).To(Succeed())
		Expect(os.Chown(grootFilePath, int(GrootUID), int(GrootGID))).To(Succeed())

		grootFolder := path.Join(sourceImagePath, "groot-folder")
		Expect(os.Mkdir(grootFolder, 0777)).To(Succeed())
		Expect(os.Chown(grootFolder, int(GrootUID), int(GrootGID))).To(Succeed())
		Expect(ioutil.WriteFile(path.Join(grootFolder, "hello"), []byte("hello-world"), 0644)).To(Succeed())

		rootFilePath := path.Join(sourceImagePath, "bar")
		Expect(ioutil.WriteFile(rootFilePath, []byte("hello-world"), 0644)).To(Succeed())

		rootFolder := path.Join(sourceImagePath, "root-folder")
		Expect(os.Mkdir(rootFolder, 0777)).To(Succeed())
		Expect(ioutil.WriteFile(path.Join(rootFolder, "hello"), []byte("hello-world"), 0644)).To(Succeed())

		grootLinkToRootFile := path.Join(sourceImagePath, "groot-link")
		Expect(os.Symlink(rootFilePath, grootLinkToRootFile)).To(Succeed())
		Expect(os.Lchown(grootLinkToRootFile, int(GrootUID), int(GrootGID)))
	})

	AfterEach(func() {
		Expect(os.RemoveAll(sourceImagePath)).To(Succeed())
		Expect(os.RemoveAll(baseImagePath)).To(Succeed())
	})

	JustBeforeEach(func() {
		baseImageFile := integration.CreateBaseImageTar(sourceImagePath)
		baseImagePath = baseImageFile.Name()
	})

	It("keeps the ownership and permissions", func() {
		integration.SkipIfNonRoot(GrootfsTestUid)

		image, err := Runner.Create(groot.CreateSpec{
			BaseImage: baseImagePath,
			ID:        "random-id",
			Mount:     true,
		})
		Expect(err).ToNot(HaveOccurred())

		grootFi, err := os.Stat(path.Join(image.RootFSPath, "foo"))
		Expect(err).NotTo(HaveOccurred())
		Expect(grootFi.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootUID)))
		Expect(grootFi.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootGID)))

		grootLink, err := os.Lstat(path.Join(image.RootFSPath, "groot-link"))
		Expect(err).NotTo(HaveOccurred())
		Expect(grootLink.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootUID)))
		Expect(grootLink.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootGID)))

		rootFi, err := os.Stat(path.Join(image.RootFSPath, "bar"))
		Expect(err).NotTo(HaveOccurred())
		Expect(rootFi.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(rootUID)))
		Expect(rootFi.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(rootGID)))
	})

	Context("when mappings are provided", func() {
		It("translates the rootfs accordingly", func() {
			image, err := Runner.WithLogLevel(lager.DEBUG).
				Create(groot.CreateSpec{
					ID:        "some-id",
					BaseImage: baseImagePath,
					Mount:     true,
					UIDMappings: []groot.IDMappingSpec{
						groot.IDMappingSpec{HostID: int(GrootUID), NamespaceID: 0, Size: 1},
						groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
					},
					GIDMappings: []groot.IDMappingSpec{
						groot.IDMappingSpec{HostID: int(GrootGID), NamespaceID: 0, Size: 1},
						groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
					},
				})

			Expect(err).NotTo(HaveOccurred())

			grootFi, err := os.Stat(path.Join(image.RootFSPath, "foo"))
			Expect(err).NotTo(HaveOccurred())
			Expect(grootFi.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootUID + 99999)))
			Expect(grootFi.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootGID + 99999)))

			grootDir, err := os.Stat(path.Join(image.RootFSPath, "groot-folder"))
			Expect(err).NotTo(HaveOccurred())
			Expect(grootDir.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootUID + 99999)))
			Expect(grootDir.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootGID + 99999)))

			grootLink, err := os.Lstat(path.Join(image.RootFSPath, "groot-link"))
			Expect(err).NotTo(HaveOccurred())
			Expect(grootLink.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootUID + 99999)))
			Expect(grootLink.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootGID + 99999)))

			rootFi, err := os.Stat(path.Join(image.RootFSPath, "bar"))
			Expect(err).NotTo(HaveOccurred())
			Expect(rootFi.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootUID)))
			Expect(rootFi.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootGID)))

			rootDir, err := os.Stat(path.Join(image.RootFSPath, "root-folder"))
			Expect(err).NotTo(HaveOccurred())
			Expect(rootDir.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootUID)))
			Expect(rootDir.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootGID)))
		})

		It("allows the mapped user to have access to the created image", func() {
			image, err := Runner.WithLogLevel(lager.DEBUG).
				Create(groot.CreateSpec{
					Mount:     true,
					ID:        "some-id",
					BaseImage: baseImagePath,
					UIDMappings: []groot.IDMappingSpec{
						groot.IDMappingSpec{HostID: int(GrootUID), NamespaceID: 0, Size: 1},
						groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
					},
					GIDMappings: []groot.IDMappingSpec{
						groot.IDMappingSpec{HostID: int(GrootGID), NamespaceID: 0, Size: 1},
						groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
					},
				})
			Expect(err).NotTo(HaveOccurred())

			listRootfsCmd := exec.Command("ls", filepath.Join(image.RootFSPath, "root-folder"))
			listRootfsCmd.SysProcAttr = &syscall.SysProcAttr{
				Credential: &syscall.Credential{
					Uid: GrootUID,
					Gid: GrootGID,
				},
			}

			sess, err := gexec.Start(listRootfsCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(gexec.Exit(0))
		})
	})

	Context("storage setup", func() {
		It("creates the storage path with the correct permission", func() {
			storePath := filepath.Join(StorePath, "new-store")
			Expect(storePath).ToNot(BeAnExistingFile())
			_, err := Runner.WithStore(storePath).Create(groot.CreateSpec{
				BaseImage: baseImagePath,
				ID:        "random-id",
				Mount:     true,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(storePath).To(BeADirectory())
			stat, err := os.Stat(storePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(stat.Mode().Perm()).To(Equal(os.FileMode(0700)))
		})

		Context("when there's no mapping", func() {
			It("sets the onwership of the store to the caller user", func() {
				_, err := Runner.WithLogLevel(lager.DEBUG).
					Create(groot.CreateSpec{
						ID:        "some-id",
						BaseImage: baseImagePath,
						Mount:     true,
					})
				Expect(err).NotTo(HaveOccurred())

				stat, err := os.Stat(filepath.Join(StorePath, store.ImageDirName))
				Expect(err).NotTo(HaveOccurred())
				Expect(stat.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootfsTestUid)))
				Expect(stat.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootfsTestGid)))

				stat, err = os.Stat(filepath.Join(StorePath, store.VolumesDirName))
				Expect(err).NotTo(HaveOccurred())
				Expect(stat.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootfsTestUid)))
				Expect(stat.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootfsTestGid)))

				stat, err = os.Stat(filepath.Join(StorePath, store.LocksDirName))
				Expect(err).NotTo(HaveOccurred())
				Expect(stat.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(GrootfsTestUid)))
				Expect(stat.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(GrootfsTestGid)))
			})
		})

		Context("when there are mappings", func() {
			It("sets the onwnership of the store to the mapped user", func() {
				_, err := Runner.WithLogLevel(lager.DEBUG).
					Create(groot.CreateSpec{
						ID:        "some-id",
						BaseImage: baseImagePath,
						Mount:     true,
						UIDMappings: []groot.IDMappingSpec{
							groot.IDMappingSpec{HostID: 1000, NamespaceID: 0, Size: 1},
							groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
						},
						GIDMappings: []groot.IDMappingSpec{
							groot.IDMappingSpec{HostID: 1000, NamespaceID: 0, Size: 1},
							groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
						},
					})

				Expect(err).ToNot(HaveOccurred())

				stat, err := os.Stat(filepath.Join(StorePath, store.ImageDirName))
				Expect(err).NotTo(HaveOccurred())
				Expect(stat.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(1000)))
				Expect(stat.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(1000)))

				stat, err = os.Stat(filepath.Join(StorePath, store.VolumesDirName))
				Expect(err).NotTo(HaveOccurred())
				Expect(stat.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(1000)))
				Expect(stat.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(1000)))

				stat, err = os.Stat(filepath.Join(StorePath, store.LocksDirName))
				Expect(err).NotTo(HaveOccurred())
				Expect(stat.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(1000)))
				Expect(stat.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(1000)))
			})

			Context("but there's no mapping for root or size = 1", func() {
				It("fails fast", func() {
					_, err := Runner.WithLogLevel(lager.DEBUG).
						Create(groot.CreateSpec{
							ID:        "some-id",
							BaseImage: baseImagePath,
							Mount:     true,
							UIDMappings: []groot.IDMappingSpec{
								groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
							},
							GIDMappings: []groot.IDMappingSpec{
								groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
							},
						})

					Expect(err).To(MatchError(ContainSubstring("couldn't determine store owner, missing root user mapping")))
				})
			})
		})

		Context("when fails to configure the store", func() {
			Describe("create", func() {
				It("logs the image id", func() {
					logBuffer := gbytes.NewBuffer()
					_, err := Runner.WithStore("/invalid/store/path").WithStderr(logBuffer).
						Create(groot.CreateSpec{
							ID:        "random-id",
							BaseImage: "my-image",
							Mount:     true,
						})
					Expect(err).To(HaveOccurred())
					Expect(logBuffer).To(gbytes.Say(`"id":"random-id"`))
				})
			})
		})
	})

	Context("when disk limit is provided", func() {
		BeforeEach(func() {
			Expect(writeMegabytes(filepath.Join(sourceImagePath, "fatfile"), 5)).To(Succeed())
		})

		It("creates a image with supplied limit", func() {
			image, err := Runner.Create(groot.CreateSpec{
				BaseImage: baseImagePath,
				ID:        "random-id",
				DiskLimit: tenMegabytes,
				Mount:     true,
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(writeMegabytes(filepath.Join(image.RootFSPath, "hello"), 4)).To(Succeed())
			Expect(writeMegabytes(filepath.Join(image.RootFSPath, "hello2"), 2)).To(MatchError(ContainSubstring("dd: error writing")))
		})

		Context("when the disk limit value is invalid", func() {
			It("fails with a helpful error", func() {
				_, err := Runner.Create(groot.CreateSpec{
					DiskLimit: -200,
					BaseImage: baseImagePath,
					ID:        "random-id",
					Mount:     true,
				})
				Expect(err).To(MatchError(ContainSubstring("disk limit cannot be negative")))
			})
		})

		Context("when the exclude-image-from-quota is also provided", func() {
			It("creates a image with supplied limit, but doesn't take into account the base image size", func() {
				image, err := Runner.Create(groot.CreateSpec{
					DiskLimit:                 10485760,
					ExcludeBaseImageFromQuota: true,
					BaseImage:                 baseImagePath,
					ID:                        "random-id",
					Mount:                     true,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(writeMegabytes(filepath.Join(image.RootFSPath, "hello"), 6)).To(Succeed())
				Expect(writeMegabytes(filepath.Join(image.RootFSPath, "hello2"), 5)).To(MatchError(ContainSubstring("dd: error writing")))
			})
		})

		Describe("--drax-bin global flag", func() {
			var (
				draxCalledFile *os.File
				draxBin        *os.File
				tempFolder     string
			)

			BeforeEach(func() {
				integration.SkipIfNotBTRFS(Driver)
				tempFolder, draxBin, draxCalledFile = integration.CreateFakeDrax()
			})

			Context("when it's provided", func() {
				It("uses the provided drax", func() {
					_, err := Runner.WithDraxBin(draxBin.Name()).Create(groot.CreateSpec{
						BaseImage: baseImagePath,
						ID:        "random-id",
						Mount:     true,
						DiskLimit: 104857600,
					})
					Expect(err).NotTo(HaveOccurred())

					contents, err := ioutil.ReadFile(draxCalledFile.Name())
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal("I'm groot - drax"))
				})

				Context("when the drax bin doesn't have uid bit set", func() {
					It("doesn't leak the image dir", func() {
						testhelpers.UnsuidDrax(draxBin.Name())
						_, err := Runner.WithDraxBin(draxBin.Name()).Create(groot.CreateSpec{
							BaseImage: baseImagePath,
							ID:        "random-id",
							Mount:     true,
							DiskLimit: 104857600,
						})
						Expect(err).To(HaveOccurred())

						imagePath := path.Join(Runner.StorePath, "images", "random-id")
						Expect(imagePath).ToNot(BeAnExistingFile())
					})
				})
			})

			Context("when it's not provided", func() {
				It("uses drax from $PATH", func() {
					newPATH := fmt.Sprintf("%s:%s", tempFolder, os.Getenv("PATH"))
					_, err := Runner.WithoutDraxBin().WithEnvVar(fmt.Sprintf("PATH=%s", newPATH)).Create(groot.CreateSpec{
						BaseImage: baseImagePath,
						ID:        "random-id",
						Mount:     true,
						DiskLimit: tenMegabytes,
					})
					Expect(err).ToNot(HaveOccurred())

					contents, err := ioutil.ReadFile(draxCalledFile.Name())
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal("I'm groot - drax"))
				})
			})
		})
	})

	Describe("unique uid and gid mappings per store", func() {
		Context("when creating two images with different mappings", func() {
			JustBeforeEach(func() {
				storePath, err := ioutil.TempDir(StorePath, "store")
				Expect(err).NotTo(HaveOccurred())
				Expect(os.Chmod(storePath, 0777)).To(Succeed())
				Runner = Runner.WithStore(storePath)

				image, err := Runner.Create(groot.CreateSpec{
					BaseImage: baseImagePath,
					ID:        "foobar",
					Mount:     true,
					UIDMappings: []groot.IDMappingSpec{
						groot.IDMappingSpec{HostID: int(GrootUID), NamespaceID: 0, Size: 1},
						groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
					},
					GIDMappings: []groot.IDMappingSpec{
						groot.IDMappingSpec{HostID: int(GrootGID), NamespaceID: 0, Size: 1},
						groot.IDMappingSpec{HostID: 100000, NamespaceID: 1, Size: 65000},
					},
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(image.Path).To(BeADirectory())
			})

			It("returns a useful error message", func() {
				_, err := Runner.Create(groot.CreateSpec{
					BaseImage: baseImagePath,
					ID:        "foobar2",
					Mount:     true,
				})
				Expect(err).To(MatchError("store already initialized with a different mapping"))
			})
		})
	})

	Context("when --with-clean is given", func() {
		BeforeEach(func() {
			_, err := Runner.Create(groot.CreateSpec{
				ID:        "my-busybox",
				BaseImage: "docker:///busybox:1.26.2",
				Mount:     true,
				DiskLimit: 10 * 1024 * 1024,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(Runner.Delete("my-busybox")).To(Succeed())
		})

		AfterEach(func() {
			Runner.Delete("my-empty")
		})

		It("cleans the store first", func() {
			preContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
			Expect(err).NotTo(HaveOccurred())
			Expect(preContents).To(HaveLen(1))

			_, err = Runner.Create(groot.CreateSpec{
				ID:            "my-empty",
				BaseImage:     "docker:///cfgarden/empty:v0.1.1",
				Mount:         true,
				CleanOnCreate: true,
			})
			Expect(err).NotTo(HaveOccurred())

			afterContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
			Expect(err).NotTo(HaveOccurred())
			Expect(afterContents).To(HaveLen(2))
			for _, layer := range testhelpers.EmptyBaseImageV011.Layers {
				Expect(filepath.Join(StorePath, store.VolumesDirName, layer.ChainID)).To(BeADirectory())
			}
		})
	})

	Context("when --without-clean is given", func() {
		BeforeEach(func() {
			_, err := Runner.Create(groot.CreateSpec{
				ID:        "my-busybox",
				BaseImage: "docker:///busybox:1.26.2",
				Mount:     true,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(Runner.Delete("my-busybox")).To(Succeed())
		})

		AfterEach(func() {
			Runner.Delete("my-empty")
		})

		It("does not clean the store", func() {
			preContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
			Expect(err).NotTo(HaveOccurred())
			Expect(preContents).To(HaveLen(1))

			_, err = Runner.Create(groot.CreateSpec{
				ID:            "my-empty",
				BaseImage:     "docker:///cfgarden/empty:v0.1.1",
				Mount:         true,
				CleanOnCreate: false,
			})
			Expect(err).NotTo(HaveOccurred())

			afterContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
			Expect(err).NotTo(HaveOccurred())
			Expect(afterContents).To(HaveLen(3))

			layers := append(testhelpers.EmptyBaseImageV011.Layers, testhelpers.BusyBoxImage.Layers...)
			for _, layer := range layers {
				Expect(filepath.Join(StorePath, store.VolumesDirName, layer.ChainID)).To(BeADirectory())
			}
		})
	})

	Context("when both without-clean and with-clean flags are given", func() {
		It("returns an error", func() {
			_, err := Runner.WithClean().WithNoClean().Create(groot.CreateSpec{
				ID:        "my-empty",
				BaseImage: "docker:///cfgarden/empty:v0.1.1",
				Mount:     true,
			})
			Expect(err).To(MatchError(ContainSubstring("with-clean and without-clean cannot be used together")))
		})
	})

	Context("when both json and no-json flags are given", func() {
		It("returns an error", func() {
			_, err := Runner.WithJson().WithNoJson().Create(groot.CreateSpec{
				ID:        "my-empty",
				BaseImage: "docker:///cfgarden/empty:v0.1.1",
				Mount:     true,
			})
			Expect(err).To(MatchError(ContainSubstring("json and no-json cannot be used together")))
		})
	})

	Context("when no --store option is given", func() {
		BeforeEach(func() {
			integration.SkipIfNotBTRFS(Driver)
			integration.SkipIfNonRoot(GrootfsTestUid)
		})

		It("uses the default store path", func() {
			Expect("/var/lib/grootfs/images").ToNot(BeAnExistingFile())
			_, err := Runner.WithoutStore().Create(groot.CreateSpec{
				BaseImage: baseImagePath,
				ID:        "random-id",
				Mount:     true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect("/var/lib/grootfs/images").To(BeADirectory())
		})
	})

	Context("when the id is already being used", func() {
		JustBeforeEach(func() {
			_, err := Runner.Create(groot.CreateSpec{
				ID:        "random-id",
				BaseImage: baseImagePath,
				Mount:     true,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails and produces a useful error", func() {
			_, err := Runner.WithStore(StorePath).Create(groot.CreateSpec{
				BaseImage: baseImagePath,
				ID:        "random-id",
				Mount:     true,
			})
			Expect(err).To(MatchError(ContainSubstring("image for id `random-id` already exists")))
		})
	})

	Context("when the id is not provided", func() {
		It("fails", func() {
			_, err := Runner.WithStore(StorePath).Create(groot.CreateSpec{
				BaseImage: baseImagePath,
				ID:        "",
				Mount:     true,
			})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the id contains invalid characters", func() {
		It("fails", func() {
			_, err := Runner.WithStore(StorePath).Create(groot.CreateSpec{
				BaseImage: baseImagePath,
				ID:        "this/is/not/okay",
				Mount:     true,
			})
			Expect(err).To(MatchError(ContainSubstring("id `this/is/not/okay` contains invalid characters: `/`")))
		})
	})

	Context("when StorePath doesn't match the given driver", func() {
		It("returns an error", func() {
			_, err := Runner.WithStore("/mnt/ext4").Create(groot.CreateSpec{
				BaseImage: baseImagePath,
				ID:        "random-id",
				Mount:     true,
			})
			Expect(err).To(MatchError("Image id 'random-id': Store path filesystem (/mnt/ext4) is incompatible with requested driver"))
		})
	})

	Context("when the requested filesystem driver is not supported", func() {
		It("fails with a useful error message", func() {
			_, err := Runner.WithDriver("dinosaurfs").Create(groot.CreateSpec{
				BaseImage: baseImagePath,
				ID:        "some-id",
				Mount:     true,
			})
			Expect(err).To(MatchError(ContainSubstring("filesystem driver not supported: dinosaurfs")))
		})
	})

	Context("when the image is invalid", func() {
		It("fails", func() {
			_, err := Runner.Create(groot.CreateSpec{
				ID:        "some-id",
				BaseImage: "*@#%^!&",
				Mount:     true,
			})
			Expect(err).To(MatchError(ContainSubstring("parsing image url: parse")))
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})
	})

	Describe("--config global flag", func() {
		var (
			cfg  config.Config
			spec groot.CreateSpec
		)

		BeforeEach(func() {
			cfg = config.Config{}
		})

		JustBeforeEach(func() {
			spec = groot.CreateSpec{
				ID:        "random-id",
				BaseImage: baseImagePath,
				Mount:     true,
			}
		})

		JustBeforeEach(func() {
			Expect(Runner.SetConfig(cfg)).To(Succeed())
		})

		Describe("store path", func() {
			BeforeEach(func() {
				var err error
				cfg.StorePath, err = ioutil.TempDir(StorePath, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(os.Chmod(cfg.StorePath, 0777)).To(Succeed())
			})

			It("uses the store path from the config file", func() {
				image, err := Runner.WithoutStore().Create(spec)
				Expect(err).NotTo(HaveOccurred())
				Expect(image.Path).To(Equal(filepath.Join(cfg.StorePath, "images/random-id")))
			})
		})

		Describe("filesystem driver", func() {
			BeforeEach(func() {
				cfg.FSDriver = "this-should-fail"
			})

			It("uses the filesystem driver from the config file", func() {
				_, err := Runner.WithoutDriver().Create(spec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("filesystem driver not supported: this-should-fail"))
			})
		})

		Describe("drax bin", func() {
			var (
				draxCalledFile *os.File
				draxBin        *os.File
				tempFolder     string
			)

			BeforeEach(func() {
				integration.SkipIfNotBTRFS(Driver)
				tempFolder, draxBin, draxCalledFile = integration.CreateFakeDrax()
				cfg.DraxBin = draxBin.Name()
			})

			It("uses the drax bin from the config file", func() {
				_, err := Runner.WithoutDraxBin().Create(groot.CreateSpec{
					BaseImage: baseImagePath,
					ID:        "random-id",
					DiskLimit: 104857600,
					Mount:     true,
				})
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadFile(draxCalledFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("I'm groot - drax"))
			})
		})

		Describe("mappings", func() {
			BeforeEach(func() {
				cfg.Create.UIDMappings = []string{"1:100000:65000", "0:1000:1"}
				cfg.Create.GIDMappings = []string{"1:100000:65000", "0:1000:1"}
			})

			It("uses the uid mappings from the config file", func() {
				image, err := Runner.Create(spec)
				Expect(err).NotTo(HaveOccurred())

				rootOwnedFile, err := os.Stat(filepath.Join(image.RootFSPath, "bar"))
				Expect(err).NotTo(HaveOccurred())
				Expect(rootOwnedFile.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(1000)))
				Expect(rootOwnedFile.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(1000)))
			})

			Context("when the UID mapping is invalid", func() {
				BeforeEach(func() {
					cfg.Create.UIDMappings = []string{"1:hello:65000", "0:1000:1"}
				})

				It("reports an error", func() {
					_, err := Runner.Create(spec)
					Expect(err).To(MatchError(ContainSubstring("parsing uid-mapping: expected integer")))
				})
			})

			Context("when the GID mapping is invalid", func() {
				BeforeEach(func() {
					cfg.Create.GIDMappings = []string{"1:hello:65000", "0:1000:1"}
				})

				It("reports an error", func() {
					_, err := Runner.Create(spec)
					Expect(err).To(MatchError(ContainSubstring("parsing gid-mapping: expected integer")))
				})
			})
		})

		Describe("disk limit size bytes", func() {
			BeforeEach(func() {
				cfg.Create.DiskLimitSizeBytes = tenMegabytes
			})

			It("creates a image with limit from the config file", func() {
				image, err := Runner.Create(spec)
				Expect(err).ToNot(HaveOccurred())

				Expect(writeMegabytes(filepath.Join(image.RootFSPath, "hello"), 11)).To(MatchError(ContainSubstring("dd: error writing")))
			})
		})

		Describe("json", func() {
			It("returns an image json as output", func() {
				image, err := Runner.Create(groot.CreateSpec{
					ID:        "random-id",
					BaseImage: baseImagePath,
					Json:      true,
					Mount:     true,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(image.ImageInfo.Rootfs).To(Equal(filepath.Join(StorePath, store.ImageDirName, "random-id/rootfs")))
				Expect(image.ImageInfo.Mount).To(BeNil())
				Expect(image.ImageInfo.Config).To(BeNil())
			})
		})

		Describe("without mount", func() {
			It("does not mount the rootfs", func() {
				image, err := Runner.Create(groot.CreateSpec{
					ID:        "some-id",
					BaseImage: baseImagePath,
					Json:      true,
					Mount:     false,
				})
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadDir(image.ImageInfo.Rootfs)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(BeEmpty())
			})

			Describe("Mount json output", func() {
				var (
					image groot.Image
				)

				JustBeforeEach(func() {
					var err error
					image, err = Runner.Create(groot.CreateSpec{
						ID:        "some-id",
						BaseImage: baseImagePath,
						Json:      true,
						Mount:     false,
					})
					Expect(err).NotTo(HaveOccurred())
				})

				Context("BTRFS", func() {
					BeforeEach(func() {
						integration.SkipIfNotBTRFS(Driver)
					})

					It("returns the mount information in the output json", func() {
						Expect(image.ImageInfo.Mount).ToNot(BeNil())
						Expect(image.ImageInfo.Mount.Destination).To(Equal(image.ImageInfo.Rootfs))
						Expect(image.ImageInfo.Mount.Type).To(Equal(""))
						Expect(image.ImageInfo.Mount.Source).To(Equal(filepath.Join(StorePath, store.ImageDirName, "some-id", "snapshot")))
						Expect(image.ImageInfo.Mount.Options).To(HaveLen(1))
						Expect(image.ImageInfo.Mount.Options[0]).To(Equal("bind"))
					})
				})

				Context("XFS", func() {
					BeforeEach(func() {
						integration.SkipIfNotXFS(Driver)
					})

					It("returns the mount information in the output json", func() {
						Expect(image.ImageInfo.Mount).ToNot(BeNil())
						Expect(image.ImageInfo.Mount.Destination).To(Equal(image.ImageInfo.Rootfs))
						Expect(image.ImageInfo.Mount.Type).To(Equal("overlay"))
						Expect(image.ImageInfo.Mount.Source).To(Equal("overlay"))
						Expect(image.ImageInfo.Mount.Options).To(HaveLen(1))
						Expect(image.ImageInfo.Mount.Options[0]).To(MatchRegexp(
							fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
								filepath.Join(StorePath, overlayxfs.LinksDirName, ".*"),
								filepath.Join(StorePath, store.ImageDirName, "some-id", overlayxfs.UpperDir),
								filepath.Join(StorePath, store.ImageDirName, "some-id", overlayxfs.WorkDir),
							),
						))
					})
				})

				Context("but `json` is not set", func() {
					It("returns an error", func() {
						_, err := Runner.Create(groot.CreateSpec{
							ID:        "my-empty",
							BaseImage: "docker:///cfgarden/empty:v0.1.1",
							Json:      false,
							Mount:     false,
						})
						Expect(err).To(MatchError(ContainSubstring("without-mount option must be used with the json option")))
					})
				})
			})
		})

		Describe("exclude image from quota", func() {
			BeforeEach(func() {
				cfg.Create.ExcludeImageFromQuota = true
				cfg.Create.DiskLimitSizeBytes = tenMegabytes
			})

			It("excludes base image from quota when config property say so", func() {
				image, err := Runner.Create(spec)
				Expect(err).ToNot(HaveOccurred())

				Expect(writeMegabytes(filepath.Join(image.RootFSPath, "hello"), 6)).To(Succeed())
				Expect(writeMegabytes(filepath.Join(image.RootFSPath, "hello2"), 5)).To(MatchError(ContainSubstring("dd: error writing")))
			})
		})
	})
})
