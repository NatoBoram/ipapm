package apt_test

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/NatoBoram/ipapm/apt"
)

const packagesExample = `Package: code
Version: 1.77.3-1681292746
Architecture: amd64
Section: devel
Priority: optional
Installed-Size: 341104
Maintainer: Microsoft Corporation <vscode-linux@microsoft.com>
Description: Code editing. Redefined.
 Visual Studio Code is a new choice of tool that combines the simplicity of
 a code editor with what developers need for the core edit-build-debug cycle.
 See https://code.visualstudio.com/docs/setup/linux for installation
 instructions and FAQ.
Homepage: https://code.visualstudio.com/
Conflicts: visual-studio-code
Depends: ca-certificates, libasound2 (>= 1.0.16), libatk-bridge2.0-0 (>= 2.5.3), libatk1.0-0 (>= 2.2.0), libatspi2.0-0 (>= 2.9.90), libc6 (>= 2.14), libc6 (>= 2.15), libc6 (>= 2.17), libc6 (>= 2.2.5), libcairo2 (>= 1.6.0), libcurl3-gnutls | libcurl3-nss | libcurl4 | libcurl3, libdbus-1-3 (>= 1.5.12), libdrm2 (>= 2.4.38), libexpat1 (>= 2.0.1), libgbm1 (>= 8.1~0), libglib2.0-0 (>= 2.16.0), libglib2.0-0 (>= 2.39.4), libgtk-3-0 (>= 3.9.10), libgtk-3-0 (>= 3.9.10) | libgtk-4-1, libnspr4 (>= 2:4.9-2~), libnss3 (>= 2:3.22), libnss3 (>= 3.26), libpango-1.0-0 (>= 1.14.0), libsecret-1-0 (>= 0.18), libx11-6, libx11-6 (>= 2:1.4.99.1), libxcb1 (>= 1.9.2), libxcomposite1 (>= 1:0.4.4-1), libxdamage1 (>= 1:1.1), libxext6, libxfixes3, libxkbcommon0 (>= 0.4.1), libxkbfile1, libxrandr2, xdg-utils (>= 1.0.2)
Recommends: libvulkan1
Provides: visual-studio-code
Replaces: visual-studio-code
SHA256: fba25ea5af751561dd4ce5ed7d972f8ed142ce3f24da2ed4b852b515819ee97c
Size: 88538392
Filename: pool/main/c/code/code_1.77.3-1681292746_amd64.deb

Package: code
Version: 1.97.0-1738713410
Architecture: amd64
Section: devel
Priority: optional
Installed-Size: 428944
Maintainer: Microsoft Corporation <vscode-linux@microsoft.com>
Description: Code editing. Redefined.
 Visual Studio Code is a new choice of tool that combines the simplicity of
 a code editor with what developers need for the core edit-build-debug cycle.
 See https://code.visualstudio.com/docs/setup/linux for installation
 instructions and FAQ.
Homepage: https://code.visualstudio.com/
Conflicts: visual-studio-code
Depends: ca-certificates, libasound2 (>= 1.0.17), libatk-bridge2.0-0 (>= 2.5.3), libatk1.0-0 (>= 2.2.0), libatspi2.0-0 (>= 2.9.90), libc6 (>= 2.14), libc6 (>= 2.16), libc6 (>= 2.17), libc6 (>= 2.2.5), libc6 (>= 2.25), libc6 (>= 2.28), libcairo2 (>= 1.6.0), libcurl3-gnutls | libcurl3-nss | libcurl4 | libcurl3, libdbus-1-3 (>= 1.9.14), libdrm2 (>= 2.4.75), libexpat1 (>= 2.1~beta3), libgbm1 (>= 17.1.0~rc2), libglib2.0-0 (>= 2.37.3), libgtk-3-0 (>= 3.9.10), libgtk-3-0 (>= 3.9.10) | libgtk-4-1, libnspr4 (>= 2:4.9-2~), libnss3 (>= 2:3.30), libnss3 (>= 3.26), libpango-1.0-0 (>= 1.14.0), libx11-6, libx11-6 (>= 2:1.4.99.1), libxcb1 (>= 1.9.2), libxcomposite1 (>= 1:0.4.4-1), libxdamage1 (>= 1:1.1), libxext6, libxfixes3, libxkbcommon0 (>= 0.5.0), libxkbfile1 (>= 1:1.1.0), libxrandr2, xdg-utils (>= 1.0.2)
Recommends: libvulkan1
Provides: visual-studio-code
Replaces: visual-studio-code
SHA256: 0f67a00c7cc406f0424d0d19e69be67eb9a2577694f300a4e44d6316af5ae16b
Size: 105236790
Filename: pool/main/c/code/code_1.97.0-1738713410_amd64.deb`

func TestParsePackages(t *testing.T) {
	packages, err := apt.ParsePackages(strings.NewReader(packagesExample))
	if err != nil {
		t.Fatalf("ParsePackages failed: %v", err)
	}

	homepageURL, err := url.Parse("https://code.visualstudio.com/")
	if err != nil {
		t.Fatalf("Failed to parse expected homepage URL: %v", err)
	}

	expected := apt.Packages{
		{
			Package:       "code",
			Version:       "1.77.3-1681292746",
			Architecture:  "amd64",
			Section:       "devel",
			Priority:      "optional",
			InstalledSize: 341104,
			Maintainer:    "Microsoft Corporation <vscode-linux@microsoft.com>",
			Description:   "Code editing. Redefined. Visual Studio Code is a new choice of tool that combines the simplicity of a code editor with what developers need for the core edit-build-debug cycle. See https://code.visualstudio.com/docs/setup/linux for installation instructions and FAQ.",
			Homepage:      homepageURL,
			Conflicts:     "visual-studio-code",
			Depends:       "ca-certificates, libasound2 (>= 1.0.16), libatk-bridge2.0-0 (>= 2.5.3), libatk1.0-0 (>= 2.2.0), libatspi2.0-0 (>= 2.9.90), libc6 (>= 2.14), libc6 (>= 2.15), libc6 (>= 2.17), libc6 (>= 2.2.5), libcairo2 (>= 1.6.0), libcurl3-gnutls | libcurl3-nss | libcurl4 | libcurl3, libdbus-1-3 (>= 1.5.12), libdrm2 (>= 2.4.38), libexpat1 (>= 2.0.1), libgbm1 (>= 8.1~0), libglib2.0-0 (>= 2.16.0), libglib2.0-0 (>= 2.39.4), libgtk-3-0 (>= 3.9.10), libgtk-3-0 (>= 3.9.10) | libgtk-4-1, libnspr4 (>= 2:4.9-2~), libnss3 (>= 2:3.22), libnss3 (>= 3.26), libpango-1.0-0 (>= 1.14.0), libsecret-1-0 (>= 0.18), libx11-6, libx11-6 (>= 2:1.4.99.1), libxcb1 (>= 1.9.2), libxcomposite1 (>= 1:0.4.4-1), libxdamage1 (>= 1:1.1), libxext6, libxfixes3, libxkbcommon0 (>= 0.4.1), libxkbfile1, libxrandr2, xdg-utils (>= 1.0.2)",
			Recommends:    "libvulkan1",
			Provides:      "visual-studio-code",
			Replaces:      "visual-studio-code",
			SHA256:        "fba25ea5af751561dd4ce5ed7d972f8ed142ce3f24da2ed4b852b515819ee97c",
			Size:          88538392,
			Filename:      "pool/main/c/code/code_1.77.3-1681292746_amd64.deb",
		},
		{
			Package:       "code",
			Version:       "1.97.0-1738713410",
			Architecture:  "amd64",
			Section:       "devel",
			Priority:      "optional",
			InstalledSize: 428944,
			Maintainer:    "Microsoft Corporation <vscode-linux@microsoft.com>",
			Description:   "Code editing. Redefined. Visual Studio Code is a new choice of tool that combines the simplicity of a code editor with what developers need for the core edit-build-debug cycle. See https://code.visualstudio.com/docs/setup/linux for installation instructions and FAQ.",
			Homepage:      homepageURL,
			Conflicts:     "visual-studio-code",
			Depends:       "ca-certificates, libasound2 (>= 1.0.17), libatk-bridge2.0-0 (>= 2.5.3), libatk1.0-0 (>= 2.2.0), libatspi2.0-0 (>= 2.9.90), libc6 (>= 2.14), libc6 (>= 2.16), libc6 (>= 2.17), libc6 (>= 2.2.5), libc6 (>= 2.25), libc6 (>= 2.28), libcairo2 (>= 1.6.0), libcurl3-gnutls | libcurl3-nss | libcurl4 | libcurl3, libdbus-1-3 (>= 1.9.14), libdrm2 (>= 2.4.75), libexpat1 (>= 2.1~beta3), libgbm1 (>= 17.1.0~rc2), libglib2.0-0 (>= 2.37.3), libgtk-3-0 (>= 3.9.10), libgtk-3-0 (>= 3.9.10) | libgtk-4-1, libnspr4 (>= 2:4.9-2~), libnss3 (>= 2:3.30), libnss3 (>= 3.26), libpango-1.0-0 (>= 1.14.0), libx11-6, libx11-6 (>= 2:1.4.99.1), libxcb1 (>= 1.9.2), libxcomposite1 (>= 1:0.4.4-1), libxdamage1 (>= 1:1.1), libxext6, libxfixes3, libxkbcommon0 (>= 0.5.0), libxkbfile1 (>= 1:1.1.0), libxrandr2, xdg-utils (>= 1.0.2)",
			Recommends:    "libvulkan1",
			Provides:      "visual-studio-code",
			Replaces:      "visual-studio-code",
			SHA256:        "0f67a00c7cc406f0424d0d19e69be67eb9a2577694f300a4e44d6316af5ae16b",
			Size:          105236790,
			Filename:      "pool/main/c/code/code_1.97.0-1738713410_amd64.deb",
		},
	}

	if len(packages) != len(expected) {
		t.Fatalf("Expected %d packages, got %d", len(expected), len(packages))
	}

	for i := range expected {
		p := packages[i]
		exp := expected[i]

		if eq := reflect.DeepEqual(p.Package, exp.Package); !eq {
			t.Errorf("[%d] Package: expected %v, got %v", i, exp.Package, p.Package)
		}
		if eq := reflect.DeepEqual(p.Version, exp.Version); !eq {
			t.Errorf("[%d] Version: expected %v, got %v", i, exp.Version, p.Version)
		}
		if eq := reflect.DeepEqual(p.Architecture, exp.Architecture); !eq {
			t.Errorf("[%d] Architecture: expected %v, got %v", i, exp.Architecture, p.Architecture)
		}
		if eq := reflect.DeepEqual(p.Section, exp.Section); !eq {
			t.Errorf("[%d] Section: expected %v, got %v", i, exp.Section, p.Section)
		}
		if eq := reflect.DeepEqual(p.Priority, exp.Priority); !eq {
			t.Errorf("[%d] Priority: expected %v, got %v", i, exp.Priority, p.Priority)
		}
		if eq := reflect.DeepEqual(p.InstalledSize, exp.InstalledSize); !eq {
			t.Errorf("[%d] InstalledSize: expected %v, got %v", i, exp.InstalledSize, p.InstalledSize)
		}
		if eq := reflect.DeepEqual(p.Maintainer, exp.Maintainer); !eq {
			t.Errorf("[%d] Maintainer: expected %v, got %v", i, exp.Maintainer, p.Maintainer)
		}
		if eq := reflect.DeepEqual(p.Description, exp.Description); !eq {
			t.Errorf("[%d] Description: expected %q, got %q", i, exp.Description, p.Description)
		}
		if eq := reflect.DeepEqual(p.Homepage, exp.Homepage); !eq {
			t.Errorf("[%d] Homepage: expected %v, got %v", i, exp.Homepage, p.Homepage)
		}
		if eq := reflect.DeepEqual(p.Conflicts, exp.Conflicts); !eq {
			t.Errorf("[%d] Conflicts: expected %v, got %v", i, exp.Conflicts, p.Conflicts)
		}
		if eq := reflect.DeepEqual(p.Depends, exp.Depends); !eq {
			t.Errorf("[%d] Depends: expected %v, got %v", i, exp.Depends, p.Depends)
		}
		if eq := reflect.DeepEqual(p.Recommends, exp.Recommends); !eq {
			t.Errorf("[%d] Recommends: expected %v, got %v", i, exp.Recommends, p.Recommends)
		}
		if eq := reflect.DeepEqual(p.Provides, exp.Provides); !eq {
			t.Errorf("[%d] Provides: expected %v, got %v", i, exp.Provides, p.Provides)
		}
		if eq := reflect.DeepEqual(p.Replaces, exp.Replaces); !eq {
			t.Errorf("[%d] Replaces: expected %v, got %v", i, exp.Replaces, p.Replaces)
		}
		if eq := reflect.DeepEqual(p.MD5sum, exp.MD5sum); !eq {
			t.Errorf("[%d] MD5sum: expected %v, got %v", i, exp.MD5sum, p.MD5sum)
		}
		if eq := reflect.DeepEqual(p.SHA1, exp.SHA1); !eq {
			t.Errorf("[%d] SHA1: expected %v, got %v", i, exp.SHA1, p.SHA1)
		}
		if eq := reflect.DeepEqual(p.SHA256, exp.SHA256); !eq {
			t.Errorf("[%d] SHA256: expected %v, got %v", i, exp.SHA256, p.SHA256)
		}
		if eq := reflect.DeepEqual(p.SHA512, exp.SHA512); !eq {
			t.Errorf("[%d] SHA512: expected %v, got %v", i, exp.SHA512, p.SHA512)
		}
		if eq := reflect.DeepEqual(p.Size, exp.Size); !eq {
			t.Errorf("[%d] Size: expected %v, got %v", i, exp.Size, p.Size)
		}
		if eq := reflect.DeepEqual(p.Filename, exp.Filename); !eq {
			t.Errorf("[%d] Filename: expected %v, got %v", i, exp.Filename, p.Filename)
		}
	}

	if t.Failed() {
		return
	}

	if eq := reflect.DeepEqual(packages, expected); !eq {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}
