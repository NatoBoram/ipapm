package apt_test

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/NatoBoram/ipapm/apt"
)

const exampleSources = `Package: rust
Format: 3.0 (native)
Binary: cargo, cargo-doc, rust-all, rustc, rustfmt, rust-clippy, rust-gdb, rust-lldb, rust-doc
Architecture: amd64 arm64 all
Version: 1.95.0~1777070268~24.04~a85c377
Maintainer: Michael Aaron Murphy <mmstick@pm.me>
Standards-Version: 4.1.1
Build-Depends: ca-certificates, curl, debhelper-compat (= 10), just, patchelf
Homepage: https://github.com/pop-os/packaging-rust
Directory: pool/noble/packaging-rust/a85c3770215da0baf36450966fcdb9d7d3b35853
Package-List:
 cargo deb devel optional arch=amd64,arm64
 cargo-doc deb devel optional arch=all
 rust-all deb devel optional arch=all
 rust-clippy deb devel optional arch=amd64,arm64
 rust-doc deb devel optional arch=all
 rust-gdb deb devel optional arch=amd64,arm64
 rust-lldb deb devel optional arch=amd64,arm64
 rustc deb devel optional arch=amd64,arm64
 rustfmt deb devel optional arch=amd64,arm64
Files:
 fc0490ad987647f742662f92ccfa99c0 2058 rust_1.95.0~1777070268~24.04~a85c377.dsc
 63f97a9f83afe57193fe91ceca4bda7a 368394800 rust_1.95.0~1777070268~24.04~a85c377.tar.xz
Checksums-Sha1:
 d6f87b48abd6d03419e66c1bb9b77c54ea71798b 2058 rust_1.95.0~1777070268~24.04~a85c377.dsc
 07fb6e58456968abb4215f0afe6a89533ce4d79a 368394800 rust_1.95.0~1777070268~24.04~a85c377.tar.xz
Checksums-Sha256:
 c5f386a9839d54181c3aacf5210b2628493d26a1dfed110192f8f259fef94a5d 2058 rust_1.95.0~1777070268~24.04~a85c377.dsc
 14caca5010e73347dc4b2ca24f35816d48bcc49bd1c4c394a90c8cb6609f540b 368394800 rust_1.95.0~1777070268~24.04~a85c377.tar.xz
Checksums-Sha512:
 3a143b678c27020842295aa4d752b49c89eb33027afb426f8312542a39690a062c900abe0005d45d2e47b94dd40537e5c1ebe6e184a390f71f5db336fa541d5f 2058 rust_1.95.0~1777070268~24.04~a85c377.dsc
 9d5f64ebf7896c675c302e8997e37c1146a952833ba4a0aa482af9dc17cb8f7051a5cbcff23dfb0fa3c797bcbc0f4fd74d0c94064a4c469e0cc211abb7288b2a 368394800 rust_1.95.0~1777070268~24.04~a85c377.tar.xz

Package: steam-installer
Format: 3.0 (native)
Binary: steam-installer, steam, steam-libs, steam-libs-i386, steam-devices
Architecture: amd64 i386 all
Version: 1:1.0.0.85~ds-2pop1~1769789104~24.04~911485f
Maintainer: Debian Games Team <pkg-games-devel@lists.alioth.debian.org>
Uploaders:  Michael Gilbert <mgilbert@debian.org>, Simon McVittie <smcv@debian.org>,
Standards-Version: 4.7.2
Build-Depends: debhelper-compat (= 12), po-debconf, python3:any, systemd-dev, zlib1g
Homepage: https://steamcommunity.com/linux
Vcs-Browser: https://salsa.debian.org/games-team/steam-installer
Vcs-Git: https://salsa.debian.org/games-team/steam-installer.git
Directory: pool/noble/steam/911485f4146588ef7aa132c9a51bb778e28e0413
Package-List:
 steam deb contrib/oldlibs optional arch=i386
 steam-devices deb games optional arch=all
 steam-installer deb contrib/games optional arch=amd64
 steam-libs deb metapackages optional arch=amd64,i386
 steam-libs-i386 deb metapackages optional arch=i386
Files:
 f9fd4cf41a18d0e56f53e6972bc542fb 2218 steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.dsc
 7a1b9ef64f6100e52e8825975398c66e 124996 steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.tar.xz
Checksums-Sha1:
 cb19481fd4962b7f12d9c66c81f72be7cff7e77f 2218 steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.dsc
 f006db342e54a48e86429ffc99e6a8c06c3ddef4 124996 steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.tar.xz
Checksums-Sha256:
 89fa9cadfd864d48dbb1aac7f80d3c5b9bd544a394576bed764c4338b21714ca 2218 steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.dsc
 ad3667777863e162cc7d8aa2b798ca2da026c3d6afbcb58dfb437d93e4c39208 124996 steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.tar.xz
Checksums-Sha512:
 39cac2bfc6c9f69b7abafee946d288631189368ead4d650f14c3b00d0b1576659ccb03d54170a0d771ca8bd9799b9c73852870460453bc1e7c405be6dbe41cad 2218 steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.dsc
 3fe889aa7de86df681da7d782f1a4bbff90e0df9a18a7a213743e4cbfbea76bcd04edc9547309cb6cbae427221863520fc23dd61b3f2adc47dab009e4dd467a4 124996 steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.tar.xz`

func TestParseSources(t *testing.T) {
	sources, err := apt.ParseSources(strings.NewReader(exampleSources))
	if err != nil {
		t.Fatalf("ParseSources failed: %v", err)
	}

	rustHomepage, err := url.Parse("https://github.com/pop-os/packaging-rust")
	if err != nil {
		t.Fatalf("Failed to parse rust homepage URL: %v", err)
	}

	steamHomepage, err := url.Parse("https://steamcommunity.com/linux")
	if err != nil {
		t.Fatalf("Failed to parse steam homepage URL: %v", err)
	}

	steamVcsBrowser, err := url.Parse("https://salsa.debian.org/games-team/steam-installer")
	if err != nil {
		t.Fatalf("Failed to parse steam vcs-browser URL: %v", err)
	}

	steamVcsGit, err := url.Parse("https://salsa.debian.org/games-team/steam-installer.git")
	if err != nil {
		t.Fatalf("Failed to parse steam vcs-git URL: %v", err)
	}

	expected := apt.Sources{
		{
			Package:          "rust",
			Format:           "3.0 (native)",
			Binary:           "cargo, cargo-doc, rust-all, rustc, rustfmt, rust-clippy, rust-gdb, rust-lldb, rust-doc",
			Architecture:     "amd64 arm64 all",
			Version:          "1.95.0~1777070268~24.04~a85c377",
			Maintainer:       "Michael Aaron Murphy <mmstick@pm.me>",
			StandardsVersion: "4.1.1",
			BuildDepends:     "ca-certificates, curl, debhelper-compat (= 10), just, patchelf",
			Homepage:         rustHomepage,
			Directory:        "pool/noble/packaging-rust/a85c3770215da0baf36450966fcdb9d7d3b35853",
			PackageList: []string{
				"cargo deb devel optional arch=amd64,arm64",
				"cargo-doc deb devel optional arch=all",
				"rust-all deb devel optional arch=all",
				"rust-clippy deb devel optional arch=amd64,arm64",
				"rust-doc deb devel optional arch=all",
				"rust-gdb deb devel optional arch=amd64,arm64",
				"rust-lldb deb devel optional arch=amd64,arm64",
				"rustc deb devel optional arch=amd64,arm64",
				"rustfmt deb devel optional arch=amd64,arm64",
			},
			Files: []apt.SourceSum{
				{Hash: "fc0490ad987647f742662f92ccfa99c0", Size: 2058, Name: "rust_1.95.0~1777070268~24.04~a85c377.dsc"},
				{Hash: "63f97a9f83afe57193fe91ceca4bda7a", Size: 368394800, Name: "rust_1.95.0~1777070268~24.04~a85c377.tar.xz"},
			},
			ChecksumsSha1: []apt.SourceSum{
				{Hash: "d6f87b48abd6d03419e66c1bb9b77c54ea71798b", Size: 2058, Name: "rust_1.95.0~1777070268~24.04~a85c377.dsc"},
				{Hash: "07fb6e58456968abb4215f0afe6a89533ce4d79a", Size: 368394800, Name: "rust_1.95.0~1777070268~24.04~a85c377.tar.xz"},
			},
			ChecksumsSha256: []apt.SourceSum{
				{Hash: "c5f386a9839d54181c3aacf5210b2628493d26a1dfed110192f8f259fef94a5d", Size: 2058, Name: "rust_1.95.0~1777070268~24.04~a85c377.dsc"},
				{Hash: "14caca5010e73347dc4b2ca24f35816d48bcc49bd1c4c394a90c8cb6609f540b", Size: 368394800, Name: "rust_1.95.0~1777070268~24.04~a85c377.tar.xz"},
			},
			ChecksumsSha512: []apt.SourceSum{
				{Hash: "3a143b678c27020842295aa4d752b49c89eb33027afb426f8312542a39690a062c900abe0005d45d2e47b94dd40537e5c1ebe6e184a390f71f5db336fa541d5f", Size: 2058, Name: "rust_1.95.0~1777070268~24.04~a85c377.dsc"},
				{Hash: "9d5f64ebf7896c675c302e8997e37c1146a952833ba4a0aa482af9dc17cb8f7051a5cbcff23dfb0fa3c797bcbc0f4fd74d0c94064a4c469e0cc211abb7288b2a", Size: 368394800, Name: "rust_1.95.0~1777070268~24.04~a85c377.tar.xz"},
			},
			Warnings: []error{},
		},
		{
			Package:          "steam-installer",
			Format:           "3.0 (native)",
			Binary:           "steam-installer, steam, steam-libs, steam-libs-i386, steam-devices",
			Architecture:     "amd64 i386 all",
			Version:          "1:1.0.0.85~ds-2pop1~1769789104~24.04~911485f",
			Maintainer:       "Debian Games Team <pkg-games-devel@lists.alioth.debian.org>",
			StandardsVersion: "4.7.2",
			BuildDepends:     "debhelper-compat (= 12), po-debconf, python3:any, systemd-dev, zlib1g",
			Homepage:         steamHomepage,
			VcsBrowser:       steamVcsBrowser,
			VcsGit:           steamVcsGit,
			Directory:        "pool/noble/steam/911485f4146588ef7aa132c9a51bb778e28e0413",
			PackageList: []string{
				"steam deb contrib/oldlibs optional arch=i386",
				"steam-devices deb games optional arch=all",
				"steam-installer deb contrib/games optional arch=amd64",
				"steam-libs deb metapackages optional arch=amd64,i386",
				"steam-libs-i386 deb metapackages optional arch=i386",
			},
			Files: []apt.SourceSum{
				{Hash: "f9fd4cf41a18d0e56f53e6972bc542fb", Size: 2218, Name: "steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.dsc"},
				{Hash: "7a1b9ef64f6100e52e8825975398c66e", Size: 124996, Name: "steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.tar.xz"},
			},
			ChecksumsSha1: []apt.SourceSum{
				{Hash: "cb19481fd4962b7f12d9c66c81f72be7cff7e77f", Size: 2218, Name: "steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.dsc"},
				{Hash: "f006db342e54a48e86429ffc99e6a8c06c3ddef4", Size: 124996, Name: "steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.tar.xz"},
			},
			ChecksumsSha256: []apt.SourceSum{
				{Hash: "89fa9cadfd864d48dbb1aac7f80d3c5b9bd544a394576bed764c4338b21714ca", Size: 2218, Name: "steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.dsc"},
				{Hash: "ad3667777863e162cc7d8aa2b798ca2da026c3d6afbcb58dfb437d93e4c39208", Size: 124996, Name: "steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.tar.xz"},
			},
			ChecksumsSha512: []apt.SourceSum{
				{Hash: "39cac2bfc6c9f69b7abafee946d288631189368ead4d650f14c3b00d0b1576659ccb03d54170a0d771ca8bd9799b9c73852870460453bc1e7c405be6dbe41cad", Size: 2218, Name: "steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.dsc"},
				{Hash: "3fe889aa7de86df681da7d782f1a4bbff90e0df9a18a7a213743e4cbfbea76bcd04edc9547309cb6cbae427221863520fc23dd61b3f2adc47dab009e4dd467a4", Size: 124996, Name: "steam-installer_1.0.0.85~ds-2pop1~1769789104~24.04~911485f.tar.xz"},
			},
			Warnings: []error{},
		},
	}

	if len(sources) != len(expected) {
		t.Fatalf("Expected %d sources, got %d", len(expected), len(sources))
	}

	for i := range expected {
		s := sources[i]
		exp := expected[i]

		if eq := reflect.DeepEqual(s.Package, exp.Package); !eq {
			t.Errorf("[%d] Package: expected %v, got %v", i, exp.Package, s.Package)
		}
		if eq := reflect.DeepEqual(s.Format, exp.Format); !eq {
			t.Errorf("[%d] Format: expected %v, got %v", i, exp.Format, s.Format)
		}
		if eq := reflect.DeepEqual(s.Binary, exp.Binary); !eq {
			t.Errorf("[%d] Binary: expected %v, got %v", i, exp.Binary, s.Binary)
		}
		if eq := reflect.DeepEqual(s.Architecture, exp.Architecture); !eq {
			t.Errorf("[%d] Architecture: expected %v, got %v", i, exp.Architecture, s.Architecture)
		}
		if eq := reflect.DeepEqual(s.Version, exp.Version); !eq {
			t.Errorf("[%d] Version: expected %v, got %v", i, exp.Version, s.Version)
		}
		if eq := reflect.DeepEqual(s.Maintainer, exp.Maintainer); !eq {
			t.Errorf("[%d] Maintainer: expected %v, got %v", i, exp.Maintainer, s.Maintainer)
		}
		if eq := reflect.DeepEqual(s.StandardsVersion, exp.StandardsVersion); !eq {
			t.Errorf("[%d] StandardsVersion: expected %v, got %v", i, exp.StandardsVersion, s.StandardsVersion)
		}
		if eq := reflect.DeepEqual(s.BuildDepends, exp.BuildDepends); !eq {
			t.Errorf("[%d] BuildDepends: expected %v, got %v", i, exp.BuildDepends, s.BuildDepends)
		}
		if eq := reflect.DeepEqual(s.Homepage, exp.Homepage); !eq {
			t.Errorf("[%d] Homepage: expected %v, got %v", i, exp.Homepage, s.Homepage)
		}
		if eq := reflect.DeepEqual(s.VcsBrowser, exp.VcsBrowser); !eq {
			t.Errorf("[%d] VcsBrowser: expected %v, got %v", i, exp.VcsBrowser, s.VcsBrowser)
		}
		if eq := reflect.DeepEqual(s.VcsGit, exp.VcsGit); !eq {
			t.Errorf("[%d] VcsGit: expected %v, got %v", i, exp.VcsGit, s.VcsGit)
		}
		if eq := reflect.DeepEqual(s.Directory, exp.Directory); !eq {
			t.Errorf("[%d] Directory: expected %v, got %v", i, exp.Directory, s.Directory)
		}
		if eq := reflect.DeepEqual(s.PackageList, exp.PackageList); !eq {
			t.Errorf("[%d] PackageList: expected %v, got %v", i, exp.PackageList, s.PackageList)
		}
		if eq := reflect.DeepEqual(s.Files, exp.Files); !eq {
			t.Errorf("[%d] Files: expected %v, got %v", i, exp.Files, s.Files)
		}
		if eq := reflect.DeepEqual(s.ChecksumsSha1, exp.ChecksumsSha1); !eq {
			t.Errorf("[%d] ChecksumsSha1: expected %v, got %v", i, exp.ChecksumsSha1, s.ChecksumsSha1)
		}
		if eq := reflect.DeepEqual(s.ChecksumsSha256, exp.ChecksumsSha256); !eq {
			t.Errorf("[%d] ChecksumsSha256: expected %v, got %v", i, exp.ChecksumsSha256, s.ChecksumsSha256)
		}
		if eq := reflect.DeepEqual(s.ChecksumsSha512, exp.ChecksumsSha512); !eq {
			t.Errorf("[%d] ChecksumsSha512: expected %v, got %v", i, exp.ChecksumsSha512, s.ChecksumsSha512)
		}
		if eq := reflect.DeepEqual(s.Warnings, exp.Warnings); !eq {
			t.Errorf("[%d] Warnings: expected %v, got %v", i, exp.Warnings, s.Warnings)
		}
	}

	if t.Failed() {
		return
	}

	if eq := reflect.DeepEqual(sources, expected); !eq {
		t.Errorf("Expected %v, got %v", expected, sources)
	}
}
