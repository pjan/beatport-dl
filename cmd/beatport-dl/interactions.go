package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"pjan/beatport-dl/config"
	"pjan/beatport-dl/internal/beatport"
)

func Setup() (cfg *config.AppConfig, cachePath string, err error) {
	configFilePath, exists, err := FindConfigFile()
	if err != nil {
		return nil, "", err
	}

	if !exists {
		fmt.Println("Config file not found, creating a new one:", configFilePath)

		fmt.Print("Username: ")
		username := GetLine()
		fmt.Print("Password: ")
		password := GetLine()
		fmt.Print("Downloads directory: ")
		downloadsDir := GetLine()

		cfg := &config.AppConfig{
			Username:           username,
			Password:           password,
			DownloadsDirectory: downloadsDir,
		}

		fmt.Println("1. Lossless (44.1 khz FLAC)\n2. High (256 kbps AAC)\n3. Medium (128 kbps AAC)\n4. Medium HLS (128 kbps AAC)")
		for {
			fmt.Print("Quality: ")
			qualityNumber := GetLine()
			switch qualityNumber {
			case "1":
				cfg.Quality = "lossless"
			case "2":
				cfg.Quality = "high"
			case "3":
				cfg.Quality = "medium"
			case "4":
				cfg.Quality = "medium-hls"
			default:
				fmt.Println("Invalid quality")
				continue
			}
			break
		}

		if err := cfg.Save(configFilePath); err != nil {
			return nil, configFilePath, fmt.Errorf("save config: %w", err)
		}
	}

	parsedConfig, err := config.Parse(configFilePath)
	if err != nil {
		return nil, configFilePath, fmt.Errorf("load config: %w", err)
	}

	cacheFilePath, exists, err := FindCacheFile()
	if err != nil {
		return nil, configFilePath, fmt.Errorf("get executable path: %w", err)
	}

	return parsedConfig, cacheFilePath, nil
}

func (app *application) mainPrompt() {
	fmt.Print("Enter url or search query: ")
	input := GetLine()
	if strings.HasPrefix(input, "https://www.beatport.com") || strings.HasPrefix(input, "https://www.beatsource.com") {
		app.urls = append(app.urls, input)
	} else {
		app.search(input)
	}
}

func (app *application) search(input string) {
	var storeTag string
	var inst *beatport.Beatport
	storeTag, input = extractStoreTag(input)
	switch storeTag {
	default:
		inst = app.bp
	case "beatsource":
		inst = app.bs
	}

	results, err := inst.Search(input)
	if err != nil {
		app.FatalError("beatport", err)
	}
	trackResultsLen := len(results.Tracks)
	releasesResultsLen := len(results.Releases)

	if trackResultsLen+releasesResultsLen == 0 {
		fmt.Println("No results found")
		return
	}

	fmt.Println("Search results:")
	fmt.Println("[ Tracks ]")
	for i, track := range results.Tracks {
		fmt.Printf(
			"%2d. %s - %s (%s) [%s]\n", i+1,
			track.Artists.Display(
				app.config.ArtistsLimit,
				app.config.ArtistsShortForm,
			),
			track.Name.String(),
			track.MixName.String(),
			track.Length,
		)
	}
	fmt.Println("\n[ Releases ]")
	indexOffset := trackResultsLen + 1
	for i, release := range results.Releases {
		fmt.Printf(
			"%2d. %s - %s [%s]\n", i+indexOffset,
			release.Artists.Display(
				app.config.ArtistsLimit,
				app.config.ArtistsShortForm,
			),
			release.Name.String(),
			release.Label.Name,
		)
	}
	fmt.Print("Enter the result number(s): ")
	input = GetLine()
	requestedResults := strings.Split(input, " ")
	for _, result := range requestedResults {
		resultInt, err := strconv.Atoi(result)
		if err != nil {
			fmt.Printf("invalid result number: %s\n", result)
			continue
		}

		if resultInt > releasesResultsLen+trackResultsLen || resultInt == 0 {
			fmt.Printf("invalid result number: %d\n", resultInt)
			continue
		}

		if resultInt >= indexOffset {
			app.urls = append(app.urls, results.Releases[resultInt-indexOffset].URL)
		} else {
			app.urls = append(app.urls, results.Tracks[resultInt-1].URL)
		}
	}
}

func extractStoreTag(query string) (store, trimmedQuery string) {
	re := regexp.MustCompile(`@\w+`)
	matches := re.FindAllString(query, -1)
	if len(matches) > 0 {
		store = strings.TrimPrefix(matches[0], "@")
		trimmedQuery = re.ReplaceAllString(query, "")
		trimmedQuery = strings.TrimSpace(trimmedQuery)
	} else {
		trimmedQuery = query
	}
	return store, trimmedQuery
}

func (app *application) parseTextFile(path string) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		app.FatalError("read input text file", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		app.urls = append(app.urls, scanner.Text())
	}
}

var (
	ErrUnsupportedLinkType  = errors.New("unsupported link type")
	ErrUnsupportedLinkStore = errors.New("unsupported link store")
)
