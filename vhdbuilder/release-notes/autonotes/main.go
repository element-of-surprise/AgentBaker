package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

/**
*** This binary autogenerates release notes for AKS VHD releases.
***
*** It accepts:
*** - a run ID from which to download artifacts.
*** - the VHD build date for output naming
*** - a comma-separated list of VHD names to include/ignore.
***
*** Examples:
*** # download ONLY 1804-gen2-gpu release notes from this run ID.
*** autonotes --build 40968951 --include 1804-gen2-gpu
***
*** # download everything EXCEPT 1804-gen2-gpu release notes from this run ID.
*** autonotes --build 40968951 --ignore 1804-gen2-gpu
***
*** # download ONLY 1604,1804,1804-containerd release notes from this run ID.
*** autonotes --build 40968951 --include 1604,1804,1804-containerd
*** # download ONLY 2019-containerd release notes from this run ID.
*** autonotes --build 76289801 --include 2019-containerd
***
*** # download everything EXCEPT 2022-containerd-gen2 release notes from this run ID.
*** autonotes --build 76289801 --ignore 2022-containerd-gen2
***
*** # download ONLY 2022-containerd,2022-containerd-gen2 release notes from this run ID.
*** autonotes --build 76289801 --include 2022-containerd,2022-containerd-gen2
**/

func main() {
	var fl flags
	flag.StringVar(&fl.build, "build", "", "run ID for the VHD build.")
	flag.StringVar(&fl.include, "include", "", "only include this list of VHD release notes.")
	flag.StringVar(&fl.ignore, "ignore", "", "ignore release notes for these VHDs")
	flag.StringVar(&fl.path, "path", defaultPath, "output path to root of VHD notes")
	flag.StringVar(&fl.date, "date", defaultDate, "date of VHD build in format YYYYMM.DD.0")

	flag.Parse()

	int := make(chan os.Signal, 1)
	signal.Notify(int, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() { <-int; cancel() }()

	if errs := run(ctx, cancel, &fl); errs != nil {
		for _, err := range errs {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

func run(ctx context.Context, cancel context.CancelFunc, fl *flags) []error {
	var include, ignore map[string]bool

	includeString := stripWhitespace(fl.include)
	if len(includeString) > 0 {
		include = map[string]bool{}
		includeTokens := strings.Split(includeString, ",")
		for _, token := range includeTokens {
			include[token] = true
		}
	}

	ignoreString := stripWhitespace(fl.ignore)
	if len(ignoreString) > 0 {
		ignore = map[string]bool{}
		ignoreTokens := strings.Split(ignoreString, ",")
		for _, token := range ignoreTokens {
			ignore[token] = true
		}
	}

	enforceInclude := len(include) > 0

	// Get windows base image versions frpm the updated windows-image.env
	var wsImageVersionFilePath = filepath.Join("vhdbuilder", "packer", "windows-image.env")
	fmt.Printf("%s\n", wsImageVersionFilePath);
	file, err := os.Open(wsImageVersionFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("line is %s\n", line)
		if strings.Contains(line, "WINDOWS_2019_BASE_IMAGE_VERSION=") {
			wsImageVersions["2019-containerd"] = strings.Split(line, "=")[1]
		} else if strings.Contains(line, "WINDOWS_2022_BASE_IMAGE_VERSION=") {
			wsImageVersions["2022-containerd"] = strings.Split(line, "=")[1]
		} else if strings.Contains(line, "WINDOWS_2022_GEN2_BASE_IMAGE_VERSION=") {
			wsImageVersions["2022-containerd-gen2"] = strings.Split(line, "=")[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	artifactsToDownload := map[string]string{}
	for key, value := range artifactToPath {
		fmt.Printf("%s - %s\n", key, value);
		if ignore[key] {
			continue
		}

		if enforceInclude && !include[key] {
			continue
		}

		artifactsToDownload[key] = value
		fmt.Printf("To download %s - %s\n", key, value);

	}

	var errc = make(chan error)
	var done = make(chan struct{})

	for sku, path := range artifactsToDownload {

		if strings.Contains(path, "AKSWindows") {
			fmt.Printf("%s\n", path);
			length := len(wsImageVersions[sku])
			var wsDate = time.Now().Format("060102")
			version := wsImageVersions[sku][0:length - 6] + wsDate
			go getReleaseNotesWindows(sku, path, fl, errc, done, version)
		} else {
			go getReleaseNotes(sku, path, fl, errc, done)
		}
	}

	var errs []error

	for i := 0; i < len(artifactsToDownload); i++ {
		select {
		case err := <-errc:
			errs = append(errs, err)
		case <-done:
			continue
		}
	}

	return errs
}

func getReleaseNotes(sku, path string, fl *flags, errc chan<- error, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	// working directory, need one per sku because the file name is
	// always "release-notes.txt" so they all overwrite each other.
	tmpdir, err := ioutil.TempDir("", "releasenotes")
	if err != nil {
		errc <- fmt.Errorf("failed to create temp working directory: %w", err)
	}
	defer os.RemoveAll(tmpdir)

	releaseNotesName := fmt.Sprintf("vhd-release-notes-%s", sku)
	releaseNotesFileIn := filepath.Join(tmpdir, "release-notes.txt")
	imageListName := fmt.Sprintf("vhd-image-bom-%s", sku)
	imageListFileIn := filepath.Join(tmpdir, "image-bom.json")

	trivyReportName := fmt.Sprintf("trivy-report-%s", sku)
	trivyReportFileIn := filepath.Join(tmpdir, "trivy-report.json")
	trivyTableName := fmt.Sprintf("trivy-images-table-%s", sku)
	trivyReportTableIn := filepath.Join(tmpdir, "trivy-images-table.txt")

	artifactsDirOut := filepath.Join(fl.path, path)
	releaseNotesFileOut := filepath.Join(artifactsDirOut, fmt.Sprintf("%s.txt", fl.date))
	imageListFileOut := filepath.Join(artifactsDirOut, fmt.Sprintf("%s-image-list.json", fl.date))

	trivyReportFileOut := filepath.Join(artifactsDirOut, fmt.Sprintf("%s-trivy-report.json", fl.date))
	trivyReportTableOut := filepath.Join(artifactsDirOut, fmt.Sprintf("%s-trivy-images-table.txt", fl.date))

	latestReleaseNotesFile := filepath.Join(artifactsDirOut, "latest.txt")
	latestImageListFile := filepath.Join(artifactsDirOut, "latest-image-list.json")

	latestTrivyReportFile := filepath.Join(artifactsDirOut, "latest-trivy-report.json")
	latestTrivyReportTable := filepath.Join(artifactsDirOut, "latest-trivy-images-table.txt")

	if err := os.MkdirAll(filepath.Dir(artifactsDirOut), 0644); err != nil {
		errc <- fmt.Errorf("failed to create parent directory %s with error: %s", artifactsDirOut, err)
		return
	}

	if err := os.MkdirAll(artifactsDirOut, 0644); err != nil {
		errc <- fmt.Errorf("failed to create parent directory %s with error: %s", artifactsDirOut, err)
		return
	}

	fmt.Printf("downloading releaseNotes '%s' from build '%s'\n", releaseNotesName, fl.build)

	cmd := exec.Command("az", "pipelines", "runs", "artifact", "download", "--run-id", fl.build, "--path", tmpdir, "--artifact-name", releaseNotesName)
	if stdout, err := cmd.CombinedOutput(); err != nil {
		if err != nil {
			errc <- fmt.Errorf("failed to download az devops releaseNotes for sku %s, err: %s, output: %s", sku, err, string(stdout))
		}
		return
	}

	if err := os.Rename(releaseNotesFileIn, releaseNotesFileOut); err != nil {
		errc <- fmt.Errorf("failed to rename file %s to %s, err: %s", releaseNotesFileIn, releaseNotesFileOut, err)
		return
	}

	data, err := os.ReadFile(releaseNotesFileOut)
	if err != nil {
		errc <- fmt.Errorf("failed to read file %s for copying, err: %s", releaseNotesFileOut, err)
	}

	err = os.WriteFile(latestReleaseNotesFile, data, 0644)
	if err != nil {
		errc <- fmt.Errorf("failed to write file %s for copying, err: %s", releaseNotesFileOut, err)
	}

	cmd = exec.Command("az", "pipelines", "runs", "artifact", "download", "--run-id", fl.build, "--path", tmpdir, "--artifact-name", imageListName)
	if stdout, err := cmd.CombinedOutput(); err != nil {
		if err != nil {
			errc <- fmt.Errorf("failed to download az devops imageList for sku %s, err: %s, output: %s", sku, err, string(stdout))
		}
		return
	}

	if err := os.Rename(imageListFileIn, imageListFileOut); err != nil {
		errc <- fmt.Errorf("failed to rename file %s to %s, err: %s", imageListFileIn, imageListFileOut, err)
		return
	}

	data, err = os.ReadFile(imageListFileOut)
	if err != nil {
		errc <- fmt.Errorf("failed to read file %s for copying, err: %s", imageListFileOut, err)
	}

	err = os.WriteFile(latestImageListFile, data, 0644)
	if err != nil {
		errc <- fmt.Errorf("failed to write file %s for copying, err: %s", latestImageListFile, err)
	}

	cmd = exec.Command("az", "pipelines", "runs", "artifact", "download", "--run-id", fl.build, "--path", tmpdir, "--artifact-name", trivyReportName)
	if stdout, err := cmd.CombinedOutput(); err != nil {
		if err != nil {
			errc <- fmt.Errorf("failed to download az devops trivy report for sku %s, err: %s, output: %s", sku, err, string(stdout))
		}
		return
	}

	if err := os.Rename(trivyReportFileIn, trivyReportFileOut); err != nil {
		errc <- fmt.Errorf("failed to rename file %s to %s, err: %s", trivyReportFileIn, trivyReportFileOut, err)
		return
	}

	data, err = os.ReadFile(trivyReportFileOut)
	if err != nil {
		errc <- fmt.Errorf("failed to read file %s for copying, err: %s", trivyReportFileOut, err)
	}

	err = os.WriteFile(latestTrivyReportFile, data, 0644)
	if err != nil {
		errc <- fmt.Errorf("failed to write file %s for copying, err: %s", latestTrivyReportFile, err)
	}

	cmd = exec.Command("az", "pipelines", "runs", "artifact", "download", "--run-id", fl.build, "--path", tmpdir, "--artifact-name", trivyTableName)
	if stdout, err := cmd.CombinedOutput(); err != nil {
		if err != nil {
			errc <- fmt.Errorf("failed to download az devops trivy report table for sku %s, err: %s, output: %s", sku, err, string(stdout))
		}
		return
	}

	if err := os.Rename(trivyReportTableIn, trivyReportTableOut); err != nil {
		errc <- fmt.Errorf("failed to rename file %s to %s, err: %s", trivyReportTableIn, trivyReportTableOut, err)
		return
	}

	data, err = os.ReadFile(trivyReportTableOut)
	if err != nil {
		errc <- fmt.Errorf("failed to read file %s for copying, err: %s", trivyReportTableOut, err)
	}

	err = os.WriteFile(latestTrivyReportTable, data, 0644)
	if err != nil {
		errc <- fmt.Errorf("failed to write file %s for copying, err: %s", latestTrivyReportTable, err)
	}
}

func getReleaseNotesWindows(sku, path string, fl *flags, errc chan<- error, done chan<- struct{}, version string) {
	defer func() { done <- struct{}{} }()

	// working directory, need one per sku because the file name is
	// always "release-notes.txt" so they all overwrite each other.
	tmpdir, err := ioutil.TempDir("", "releasenotes")
	if err != nil {
		errc <- fmt.Errorf("failed to create temp working directory: %w", err)
	}
	defer os.RemoveAll(tmpdir)

	releaseNotesName := fmt.Sprintf("vhd-release-notes-%s", sku)
	releaseNotesFileIn := filepath.Join(tmpdir, "release-notes.txt")
	imageListName := fmt.Sprintf("vhd-image-bom-%s", sku)
	imageListFileIn := filepath.Join(tmpdir, "image-bom.json")

	artifactsDirOut := filepath.Join(fl.path, path)
	releaseNotesFileOut := filepath.Join(artifactsDirOut, fmt.Sprintf("%s.txt", version))
	imageListFileOut := filepath.Join(artifactsDirOut, fmt.Sprintf("%s-image-list.json", version))
	
	if err := os.MkdirAll(filepath.Dir(artifactsDirOut), 0644); err != nil {
		errc <- fmt.Errorf("failed to create parent directory %s with error: %s", artifactsDirOut, err)
		return
	}

	if err := os.MkdirAll(artifactsDirOut, 0644); err != nil {
		errc <- fmt.Errorf("failed to create parent directory %s with error: %s", artifactsDirOut, err)
		return
	}
	
	fmt.Printf("downloading releaseNotes '%s' from build '%s'\n", releaseNotesName, fl.build)

	cmd := exec.Command("az", "pipelines", "runs", "artifact", "download", "--run-id", fl.build, "--path", tmpdir, "--artifact-name", releaseNotesName)
	if stdout, err := cmd.CombinedOutput(); err != nil {
		if err != nil {
			errc <- fmt.Errorf("failed to download az devops releaseNotes for sku %s, err: %s, output: %s", sku, err, string(stdout))
		}
		return
	}

	if err := os.Rename(releaseNotesFileIn, releaseNotesFileOut); err != nil {
		errc <- fmt.Errorf("failed to rename file %s to %s, err: %s", releaseNotesFileIn, releaseNotesFileOut, err)
		return
	}

	cmd = exec.Command("az", "pipelines", "runs", "artifact", "download", "--run-id", fl.build, "--path", tmpdir, "--artifact-name", imageListName)
	if stdout, err := cmd.CombinedOutput(); err != nil {
		if err != nil {
			errc <- fmt.Errorf("failed to download az devops imageList for sku %s, err: %s, output: %s", sku, err, string(stdout))
		}
		return
	}

	if err := os.Rename(imageListFileIn, imageListFileOut); err != nil {
		errc <- fmt.Errorf("failed to rename file %s to %s, err: %s", imageListFileIn, imageListFileOut, err)
		return
	}
}

func stripWhitespace(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

type flags struct {
	build   string
	include string // CSV of the map keys below.
	ignore  string // CSV of the map keys below.
	path    string // output path
	date    string // date of vhd build
}

var defaultPath = filepath.Join("vhdbuilder", "release-notes")
var defaultDate = strings.Split(time.Now().Format("200601.02"), " ")[0] + ".0"
var wsImageVersions = make(map[string]string)
	
// why does ubuntu use subfolders and mariner doesn't
// there are dependencies on the folder structure but it would
// be nice to fix this.
var artifactToPath = map[string]string{
	"1804-containerd":                   filepath.Join("AKSUbuntu", "gen1", "1804containerd"),
	"1804-gen2-containerd":              filepath.Join("AKSUbuntu", "gen2", "1804containerd"),
	"1804-gpu-containerd":               filepath.Join("AKSUbuntu", "gen1", "1804gpucontainerd"),
	"1804-gen2-gpu-containerd":          filepath.Join("AKSUbuntu", "gen2", "1804gpucontainerd"),
	"1804-fips-containerd":              filepath.Join("AKSUbuntu", "gen1", "1804fipscontainerd"),
	"1804-fips-gen2-containerd":         filepath.Join("AKSUbuntu", "gen2", "1804fipscontainerd"),
	"2004-fips-containerd":              filepath.Join("AKSUbuntu", "gen1", "2004fipscontainerd"),
	"2004-fips-gen2-containerd":         filepath.Join("AKSUbuntu", "gen2", "2004fipscontainerd"),
	"marinerv1":                         filepath.Join("AKSCBLMariner", "gen1"),
	"marinerv1-gen2":                    filepath.Join("AKSCBLMariner", "gen2"),
	"marinerv2-gen1-fips":               filepath.Join("AKSCBLMarinerV2", "gen1fips"),
	"marinerv2-gen2-fips":               filepath.Join("AKSCBLMarinerV2", "gen2fips"),
	"marinerv2-gen2":                    filepath.Join("AKSCBLMarinerV2", "gen2"),
	"marinerv2-gen2-kata":               filepath.Join("AKSCBLMarinerV2", "gen2kata"),
	"marinerv2-gen2-arm64":              filepath.Join("AKSCBLMarinerV2", "gen2arm64"),
	"marinerv2-gen2-trustedlaunch":      filepath.Join("AKSCBLMarinerV2", "gen2tl"),
	"marinerv2-gen2-kata-trustedlaunch": filepath.Join("AKSCBLMarinerV2", "gen2katatl"),
	"2004-cvm-gen2-containerd":          filepath.Join("AKSUbuntu", "gen2", "2004cvmcontainerd"),
	"2204-containerd":                   filepath.Join("AKSUbuntu", "gen1", "2204containerd"),
	"2204-gen2-containerd":              filepath.Join("AKSUbuntu", "gen2", "2204containerd"),
	"2204-arm64-gen2-containerd":        filepath.Join("AKSUbuntu", "gen2", "2204arm64containerd"),
	"2204-tl-gen2-containerd":           filepath.Join("AKSUbuntu", "gen2", "2204tlcontainerd"),
	"2019-containerd":                filepath.Join("AKSWindows", "2019-containerd"),
	"2022-containerd":                filepath.Join("AKSWindows", "2022-containerd"),
	"2022-containerd-gen2":           filepath.Join("AKSWindows", "2022-containerd-gen2"),	
}
