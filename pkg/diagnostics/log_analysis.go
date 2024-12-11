package diagnostics

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spf13/cobra"
)

// LogAnalysisOptions holds parameters for log analysis
type LogAnalysisOptions struct {
	verbose			      bool
	tailLines             int64
	fuzzyMatch            bool
	searchString          string
	noobaaTimestamp       bool
	matchWholeString      bool
	caseInsensitivity     bool
	showLineIfNoTimestamp bool
}

const (
	noobaaOutputTimestampFormat = "02/01/2006 15:04:05"
	// To be used in case of additional NooBaa operators in the future
	noobaaOperatorPodNamePrefix = "noobaa-operator"
	operatorTimestampPrefix     = `time="`
	operatorWTimestampPrefix    = "W"
)

var (
	ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// RunLogAnalysis runs a CLI command
func RunLogAnalysis(cmd *cobra.Command, args []string) {
	verbose, _ := cmd.Flags().GetBool("verbose")
	tailLines, _ := cmd.Flags().GetInt64("tail")
	fuzzyMatch, _ := cmd.Flags().GetBool("fuzzy")
	noobaaTimestamp, _ := cmd.Flags().GetBool("noobaa-time")
	matchWholeString, _ := cmd.Flags().GetBool("whole-string")
	showLineIfNoTimestamp, _ := cmd.Flags().GetBool("prefer-line")
	caseInsensitivity, _ := cmd.Flags().GetBool("case-insensitive")
	searchString := ""

	analysisOptions := LogAnalysisOptions{
		verbose:               verbose,
		tailLines:             tailLines,
		fuzzyMatch:            fuzzyMatch,
		searchString:          searchString,
		noobaaTimestamp:       noobaaTimestamp,
		matchWholeString:      matchWholeString,
		caseInsensitivity:     caseInsensitivity,
		showLineIfNoTimestamp: showLineIfNoTimestamp,
	}

	validateLogAnalysisParameters(&analysisOptions, args)
	podSelector, _ := labels.Parse("app=noobaa")
	listOptions := client.ListOptions{Namespace: options.Namespace, LabelSelector: podSelector}
	CollectAndAnalyzeLogs(listOptions, &analysisOptions)
}

// validateLogAnalysisParameters validates the parameters for log analysis
func validateLogAnalysisParameters(analysisOptions *LogAnalysisOptions, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		analysisOptions.searchString = util.ShowStringPrompt("Provide a search string")
	} else {
		analysisOptions.searchString = args[0]
	}

	if analysisOptions.tailLines < 1 {
		log.Fatalf("❌ Tail must be a whole positive number")
	}

	if analysisOptions.fuzzyMatch && analysisOptions.matchWholeString {
		log.Fatalf("❌ Cannot use both fuzzy matching and whole-string matching")
	}
}

// CollectAndAnalyzeLogs collects and analyzes logs of all existing noobaa pods
func CollectAndAnalyzeLogs(listOptions client.ListOptions, analysisOptions *LogAnalysisOptions) {
	log := util.Logger()
	chosenTimestamp := ""
	if analysisOptions.noobaaTimestamp {
		chosenTimestamp = "NooBaa"
	} else {
		chosenTimestamp = "Kubernetes"
	}
	log.Println()
	log.Println("✨────────────────────────────────────────────✨")
	log.Printf("   Collecting and analyzing pod logs -")
	log.Printf("   Search string: %s", analysisOptions.searchString)
	log.Printf("   Case insensitivity: %t", analysisOptions.caseInsensitivity)
	log.Printf("   Match whole string: %t", analysisOptions.matchWholeString)
	log.Printf("   From the last %d lines", analysisOptions.tailLines)
	log.Printf("   Using %s timestamps", chosenTimestamp)
	log.Println("   Found occurrences will be printed below")
	log.Println("   in the format <pod name>:<container name>")
	log.Println("✨────────────────────────────────────────────✨")
	podList := &corev1.PodList{}
	if !util.KubeList(podList, &listOptions) {
		log.Printf(`❌ failed to get NooBaa pod list within namespace %s\n`, options.Namespace)
		return
	}
	for i := range podList.Items {
		pod := &podList.Items[i]
		analyzePodLogs(pod, analysisOptions)
	}
}

// analyzePodLogs will count the number of occurrences of the search string in a pod log
// as well as find and print the timestamps of the first and last occurrence of
// the search string in the logs
func analyzePodLogs(pod *corev1.Pod, analysisOptions *LogAnalysisOptions) {
	log := util.Logger()
	podLogs, err := util.GetPodLogs(*pod, &analysisOptions.tailLines, true)
	if err != nil {
		log.Printf("❌ Failed to get logs for pod %s: %v", pod.Name, err)
		return
	}
	if analysisOptions.caseInsensitivity {
		analysisOptions.searchString = strings.ToLower(analysisOptions.searchString)
	}
	stringBoundaryRegex := compileStringBoundaryRegex(analysisOptions) // Compiled here for better efficiency
	for containerName, containerLog := range podLogs {
		firstAppearanceFound := false
		firstAppearanceTimestamp := ""
		lastAppearanceTimestamp := ""
		lastOccurrenceLine := ""
		log.Printf("Analyzing %s:%s", pod.Name, containerName)
		defer containerLog.Close()
		occurrenceCounter := 0
		scanner := bufio.NewScanner(containerLog)
		for scanner.Scan() {
			line := scanner.Text()
			// Clean line from ANSI escape codes
			if !strings.Contains(pod.Name, noobaaOperatorPodNamePrefix) {
				line = sanitizeANSI(line)
			}
			lineContainsMatch := stringMatchCheck(line, stringBoundaryRegex, analysisOptions)
			if lineContainsMatch {
				if !firstAppearanceFound {
					firstAppearanceFound = true
					firstAppearanceTimestamp = extractTimeString(pod, line, *analysisOptions)
				}
				if analysisOptions.verbose {
					log.Println(line)
				}
				occurrenceCounter++
				lastOccurrenceLine = line
			}
		}
		lastAppearanceTimestamp = extractTimeString(pod, lastOccurrenceLine, *analysisOptions)
		if occurrenceCounter == 0 {
			log.Println("No occurrences found")
		} else {
			log.Printf("Hits: %d", occurrenceCounter)
			log.Printf("Earliest appearance: %s", firstAppearanceTimestamp)
			if occurrenceCounter > 1 {
				log.Printf("Latest appearance:   %s", lastAppearanceTimestamp)
			}
		}
		log.Println("──────────────────────────────────────────────────────────────────────────────────")
	}
}

// sanitizeANSI removes ANSI escape codes from a string
func sanitizeANSI(line string) string {
	// This is done in order to avoid the terminal from interpreting 
	// them as color codes and printing them as garbage characters.
	return ansiRegex.ReplaceAllString(line, "")
}

// compileStringBoundaryRegex compiles a word boundary regex pattern for the search string
func compileStringBoundaryRegex(analysisOptions *LogAnalysisOptions) *regexp.Regexp {
	var stringBoundarySearchPattern *regexp.Regexp
	stringBoundaryPattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(analysisOptions.searchString))
	if analysisOptions.caseInsensitivity {
		stringBoundarySearchPattern = regexp.MustCompile("(?i)" + stringBoundaryPattern)
	} else {
		stringBoundarySearchPattern = regexp.MustCompile(stringBoundaryPattern)
	}
	return stringBoundarySearchPattern
}

// stringMatchCheck checks if a line contains a match to the search string
func stringMatchCheck(line string, stringBoundaryRegex *regexp.Regexp, analysisOptions *LogAnalysisOptions) bool {
	if analysisOptions.matchWholeString {
		return wholestringMatchCheck(line, stringBoundaryRegex)
	} else {
		return partialMatchCheck(line, analysisOptions)
	}
}

// wholestringMatchCheck checks if a line contains a whole string match to the search string
// Mostly used for readability and organization purposes
func wholestringMatchCheck(line string, stringBoundaryRegex *regexp.Regexp) bool {
	return stringBoundaryRegex.MatchString(line)
}

// partialMatchCheck checks if a line contains a partial/fuzzy match to the search string
func partialMatchCheck(line string, analysisOptions *LogAnalysisOptions) bool {
	if analysisOptions.fuzzyMatch {
		fuzzyCaseInsensitiveMatch := analysisOptions.caseInsensitivity && fuzzy.MatchNormalized(analysisOptions.searchString, line)
		fuzzyCaseSensitiveMatch := fuzzy.Match(analysisOptions.searchString, line)
		return fuzzyCaseInsensitiveMatch || fuzzyCaseSensitiveMatch
	} else {
		// Check for a match by temporarily casting the line string to lowercase
		// (the search string is cast in the beginning of analyzePodLogs)
		caseInsensitiveMatch := analysisOptions.caseInsensitivity && strings.Contains(strings.ToLower(line), analysisOptions.searchString)
		caseSensitiveMatch := strings.Contains(line, analysisOptions.searchString)
		return caseInsensitiveMatch || caseSensitiveMatch
	}
}

// extractTimeString extracts the timestamp from a log line by checking which pod
// it originated from and redirecting it to the appropriate extraction function
func extractTimeString(pod *corev1.Pod, line string, analysisOptions LogAnalysisOptions) string {
	if analysisOptions.noobaaTimestamp {
		if strings.Contains(pod.Name, noobaaOperatorPodNamePrefix) {
			return extractOperatorTimestampString(line, analysisOptions.showLineIfNoTimestamp)
		} else {
			return extractCoreTimestampString(line, analysisOptions.showLineIfNoTimestamp)
		}
	} else {
		return extractKubernetesTimestampString(line)
	}
}

// extractKubernetesTimestampString extracts the timestamp from a Kubernetes log line
func extractKubernetesTimestampString(line string) string {
	// Example log line:
	// 2024-12-10T07:27:16.856641898Z Dec-10 7:27:16.847 [BGWorkers/36]...
	splitLine := strings.SplitN(line, " ", 2)
	return splitLine[0]
}

// extractCoreTimestampString extracts, parses and formats a timestamp string
// from pods running NooBaa Core code (core, endpoint, PV pod)
func extractCoreTimestampString(line string, showLineIfNoTimestamp bool) string {
	// Example log line:
	// Dec-9 15:16:31.621 [BGWorkers/36] [L0] ...
	const minimumRequiredIndices = 2
	// Example split result:
	// ["Dec-9", "15:16:31.621", "[BGWorkers..."]
	splitLine := strings.SplitN(line, " ", 3)
	if len(splitLine) < minimumRequiredIndices {
		return timeParsingError(showLineIfNoTimestamp, line)
	}
	lineDate := splitLine[0]
	lineTime := splitLine[1]
	// The year is assumed to be the current one since it's not provided
	year := time.Now().Year()
	layout := "Jan-2-2006 15:04:05.000"
	timestampString := fmt.Sprintf("%s-%d %s", lineDate, year, lineTime)
	parsedTime, err := time.Parse(layout, timestampString)
	if err != nil {
		return timeParsingError(showLineIfNoTimestamp, line)
	}
	return parsedTime.Format(noobaaOutputTimestampFormat)
}

// extractOperatorTimestampString extracts, parses, formats and returns a timestamp
// string from the NooBaa Operator pod logs
func extractOperatorTimestampString(line string, showLineIfNoTimestamp bool) string {
	if strings.HasPrefix(line, operatorTimestampPrefix) {
		return extractStandardOperatorTimestampString(line, showLineIfNoTimestamp)
	} else if strings.HasPrefix(line, operatorWTimestampPrefix) {
		return extractOperatorWTimestampString(line, showLineIfNoTimestamp)
	}
	return timeParsingError(showLineIfNoTimestamp, line)
}

// extractStandardOperatorTimestampString extracts the timestamp in case of a standard operator log line
// Example:
// time="2024-12-10T07:27:36Z" level=info msg="...
func extractStandardOperatorTimestampString(line string, showLineIfNoTimestamp bool) string {
	secondQuotesIndex := strings.Index(line[len(operatorTimestampPrefix):], `"`)
	if secondQuotesIndex == -1 {
		return timeParsingError(showLineIfNoTimestamp, line)
	}
	timestampString := line[len(operatorTimestampPrefix) : len(operatorTimestampPrefix)+secondQuotesIndex]
	// Parse the date using RFC3339 layout
	const operatorTimestampFormat = time.RFC3339
	parsedTimestamp, err := time.Parse(operatorTimestampFormat, timestampString)
	if err != nil {
		return timeParsingError(showLineIfNoTimestamp, line)
	}
	return parsedTimestamp.Format(noobaaOutputTimestampFormat)
}

// extractOperatorWTimestampString extracts the timestamp in case of a non-standard operator log line
// Example:
// W1209 13:41:05.890285 1 reflector.go:484...
func extractOperatorWTimestampString(line string, showLineIfNoTimestamp bool) string {
	const minimumWTimeParseLength = 22
	if len(line) < minimumWTimeParseLength {
		return timeParsingError(showLineIfNoTimestamp, line)
	}
	wStringDayMonthStartIndex := 1
	wStringDayMonthEndIndex := 5
	wStringTimeStartIndex := 6
	wStringTimeEndIndex := 21
	datePart := line[wStringDayMonthStartIndex:wStringDayMonthEndIndex]
	timePart := line[wStringTimeStartIndex:wStringTimeEndIndex]
	day := datePart[2:]
	month := datePart[:2]
	// The year is assumed to be the current one since it's not provided
	year := time.Now().Year()
	fullTimeStr := fmt.Sprintf("%s/%s/%d %s", day, month, year, timePart)
	return fullTimeStr
}

// timeParsingError returns an error message if the timestamp
// could not be parsed based on the user's prefer-line setting
func timeParsingError(showLineIfNoTimestamp bool, line string) string {
	absentTimestampErrorWithLine := fmt.Sprintf("Could not parse timestamp in line %s", line)
	const absentTimestampError = "No timestamp found"
	if showLineIfNoTimestamp {
		return absentTimestampErrorWithLine
	}
	return absentTimestampError
}
