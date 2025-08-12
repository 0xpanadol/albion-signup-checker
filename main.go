package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// Player represents a guild member
type Player struct {
	Username string
	Status   string
	Roles    string
}

// AlternativeNames holds mappings from guild names to alternative names
type AlternativeNames struct {
	GuildToAlternatives map[string][]string // guild name -> list of alternative names
	AlternativeToGuild  map[string]string   // alternative name -> guild name
}

// MatchResult represents the result of a name matching operation
type MatchResult struct {
	Found           bool
	GuildName       string
	AlternativeName string
	MatchType       string // "direct", "alternative", "ignored"
}

// parseGuildFile reads and parses the guild.txt file
func parseGuildFile(filename string) ([]Player, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open guild file: %w", err)
	}
	defer file.Close()

	var players []Player
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and header
		if line == "" || lineNum == 1 {
			continue
		}

		player, err := parseGuildLine(line)
		if err != nil {
			log.Printf("Warning: skipping malformed line %d: %v", lineNum, err)
			continue
		}

		players = append(players, player)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading guild file: %w", err)
	}

	return players, nil
}

// parseGuildLine parses a single line from guild.txt
func parseGuildLine(line string) (Player, error) {
	// Split by tabs
	parts := strings.Split(line, "\t")
	if len(parts) < 3 {
		return Player{}, fmt.Errorf("expected 3 tab-separated fields, got %d", len(parts))
	}

	// Extract quoted fields
	username, err := extractQuotedField(parts[0])
	if err != nil {
		return Player{}, fmt.Errorf("invalid username field: %w", err)
	}

	status, err := extractQuotedField(parts[1])
	if err != nil {
		return Player{}, fmt.Errorf("invalid status field: %w", err)
	}

	roles, err := extractQuotedField(parts[2])
	if err != nil {
		return Player{}, fmt.Errorf("invalid roles field: %w", err)
	}

	return Player{
		Username: username,
		Status:   status,
		Roles:    roles,
	}, nil
}

// parseAlternativeNamesFile reads and parses the sheet-names.txt file
func parseAlternativeNamesFile(filename string) (*AlternativeNames, error) {
	altNames := &AlternativeNames{
		GuildToAlternatives: make(map[string][]string),
		AlternativeToGuild:  make(map[string]string),
	}

	file, err := os.Open(filename)
	if err != nil {
		// File doesn't exist, return empty mappings
		return altNames, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse line: GuildName:AlternativeName1,AlternativeName2
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			log.Printf("Warning: skipping malformed alternative name line: %s", line)
			continue
		}

		guildName := strings.TrimSpace(parts[0])
		alternativesStr := strings.TrimSpace(parts[1])

		if guildName == "" || alternativesStr == "" {
			continue
		}

		// Parse alternative names
		alternatives := strings.Split(alternativesStr, ",")
		for _, alt := range alternatives {
			alt = strings.TrimSpace(alt)
			if alt != "" {
				// Store both directions of mapping (case-insensitive)
				altNames.GuildToAlternatives[guildName] = append(altNames.GuildToAlternatives[guildName], alt)
				altNames.AlternativeToGuild[strings.ToLower(alt)] = guildName
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading alternative names file: %w", err)
	}

	return altNames, nil
}

// extractQuotedField extracts content from a quoted field
func extractQuotedField(field string) (string, error) {
	field = strings.TrimSpace(field)
	if len(field) < 2 {
		return "", fmt.Errorf("field too short to contain quotes")
	}

	if !strings.HasPrefix(field, "\"") || !strings.HasSuffix(field, "\"") {
		return "", fmt.Errorf("field is not properly quoted")
	}

	// Remove surrounding quotes
	return field[1 : len(field)-1], nil
}

// parseSheetFile reads and parses the sheet.txt file
func parseSheetFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open sheet file: %w", err)
	}
	defer file.Close()

	var names []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Clean the name (remove parentheses content and extra spaces)
		cleanName := cleanPlayerName(line)
		if cleanName != "" {
			names = append(names, cleanName)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading sheet file: %w", err)
	}

	return names, nil
}

// cleanPlayerName removes parentheses content and normalizes the name
func cleanPlayerName(name string) string {
	// Remove content in parentheses (e.g., "(realm)", "(Longbow)")
	re := regexp.MustCompile(`\s*\([^)]*\)\s*`)
	cleaned := re.ReplaceAllString(name, "")

	// Remove extra whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Skip obviously invalid entries
	if strings.Contains(strings.ToLower(cleaned), "delete") ||
		strings.Contains(strings.ToLower(cleaned), "spam") ||
		strings.Contains(strings.ToLower(cleaned), "mess") ||
		strings.Contains(strings.ToLower(cleaned), "pedo") {
		return ""
	}

	return cleaned
}

// findNameMatch checks if a guild name exists in the sheet names, using alternative names
func findNameMatch(guildName string, sheetNames []string, altNames *AlternativeNames, ignoredNames []string) MatchResult {
	guildNameLower := strings.ToLower(guildName)

	// Check direct match first
	for _, sheetName := range sheetNames {
		if strings.ToLower(sheetName) == guildNameLower {
			return MatchResult{
				Found:     true,
				GuildName: guildName,
				MatchType: "direct",
			}
		}
	}

	// Check alternative names
	if alternatives, exists := altNames.GuildToAlternatives[guildName]; exists {
		for _, alt := range alternatives {
			altLower := strings.ToLower(alt)
			for _, sheetName := range sheetNames {
				if strings.ToLower(sheetName) == altLower {
					return MatchResult{
						Found:           true,
						GuildName:       guildName,
						AlternativeName: alt,
						MatchType:       "alternative",
					}
				}
			}
		}
	}

	// Check ignored patterns (legacy support)
	for _, sheetName := range sheetNames {
		for _, ignored := range ignoredNames {
			ignoredLower := strings.ToLower(ignored)
			if strings.Contains(guildNameLower, ignoredLower) && strings.Contains(strings.ToLower(sheetName), ignoredLower) {
				return MatchResult{
					Found:           true,
					GuildName:       guildName,
					AlternativeName: sheetName,
					MatchType:       "ignored",
				}
			}
		}
	}

	return MatchResult{Found: false}
}

// findSheetNameMatch checks if a sheet name exists in guild names, using alternative names
func findSheetNameMatch(sheetName string, guildNames []string, altNames *AlternativeNames, ignoredNames []string) MatchResult {
	sheetNameLower := strings.ToLower(sheetName)

	// Check direct match first
	for _, guildName := range guildNames {
		if strings.ToLower(guildName) == sheetNameLower {
			return MatchResult{
				Found:     true,
				GuildName: guildName,
				MatchType: "direct",
			}
		}
	}

	// Check if sheet name is an alternative name
	if guildName, exists := altNames.AlternativeToGuild[sheetNameLower]; exists {
		// Verify the guild name actually exists in the guild list
		for _, name := range guildNames {
			if name == guildName {
				return MatchResult{
					Found:           true,
					GuildName:       guildName,
					AlternativeName: sheetName,
					MatchType:       "alternative",
				}
			}
		}
	}

	// Check ignored patterns (legacy support)
	for _, guildName := range guildNames {
		for _, ignored := range ignoredNames {
			ignoredLower := strings.ToLower(ignored)
			if strings.Contains(sheetNameLower, ignoredLower) && strings.Contains(strings.ToLower(guildName), ignoredLower) {
				return MatchResult{
					Found:           true,
					GuildName:       guildName,
					AlternativeName: sheetName,
					MatchType:       "ignored",
				}
			}
		}
	}

	return MatchResult{Found: false}
}

// getExcludedRoles returns a list of roles that should be excluded from results
func getExcludedRoles() []string {
	return []string{
		"Bomber",
		"Guild Master",
	}
}

// getIgnoredNames returns a list of names/partial names that should be ignored in matching
func getIgnoredNames() []string {
	return []string{
		"sarge",
	}
}

// hasExcludedRole checks if a player has any of the excluded roles
func hasExcludedRole(playerRoles string, excludedRoles []string) bool {
	if playerRoles == "" {
		return false
	}

	// Split roles by semicolon
	roles := strings.Split(playerRoles, ";")

	// Create a map of player's roles for quick lookup (case-insensitive)
	playerRoleMap := make(map[string]bool)
	for _, role := range roles {
		cleanRole := strings.TrimSpace(role)
		if cleanRole != "" {
			playerRoleMap[strings.ToLower(cleanRole)] = true
		}
	}

	// Check if any excluded role is present
	for _, excludedRole := range excludedRoles {
		if playerRoleMap[strings.ToLower(excludedRole)] {
			return true
		}
	}

	return false
}

// findOnlinePlayersNotInSheet finds players who are online but not in the sheet and don't have excluded roles
func findOnlinePlayersNotInSheet(guildPlayers []Player, sheetNames []string, altNames *AlternativeNames) ([]string, []string, []MatchResult) {
	excludedRoles := getExcludedRoles()
	ignoredNames := getIgnoredNames()
	var result []string
	var excluded []string
	var matches []MatchResult

	for _, player := range guildPlayers {
		// Check if player is online
		if player.Status == "Online" {
			// Check if player is NOT in sheet (using improved name matching)
			matchResult := findNameMatch(player.Username, sheetNames, altNames, ignoredNames)
			if !matchResult.Found {
				// Check if player has excluded roles
				if hasExcludedRole(player.Roles, excludedRoles) {
					excluded = append(excluded, player.Username)
				} else {
					result = append(result, player.Username)
				}
			} else {
				// Player was found in sheet, record the match
				matches = append(matches, matchResult)
			}
		}
	}

	return result, excluded, matches
}

// findSheetPlayersNotInGuild finds players who are in the sheet but not in the guild
func findSheetPlayersNotInGuild(guildPlayers []Player, sheetNames []string, altNames *AlternativeNames) ([]string, []MatchResult) {
	ignoredNames := getIgnoredNames()
	var result []string
	var matches []MatchResult

	// Create list of all guild player names
	var guildNames []string
	for _, player := range guildPlayers {
		guildNames = append(guildNames, player.Username)
	}

	for _, sheetName := range sheetNames {
		// Check if sheet player is NOT in guild (using improved name matching)
		matchResult := findSheetNameMatch(sheetName, guildNames, altNames, ignoredNames)
		if !matchResult.Found {
			result = append(result, sheetName)
		} else {
			// Sheet player was found in guild, record the match
			matches = append(matches, matchResult)
		}
	}

	return result, matches
}

func main() {
	// Parse alternative names file
	fmt.Println("Loading alternative name mappings...")
	altNames, err := parseAlternativeNamesFile("data/sheet-names.txt")
	if err != nil {
		log.Fatalf("Error parsing alternative names file: %v", err)
	}
	fmt.Printf("Loaded %d alternative name mappings\n", len(altNames.GuildToAlternatives))

	// Parse guild file
	fmt.Println("Reading guild data...")
	guildPlayers, err := parseGuildFile("data/guild.txt")
	if err != nil {
		log.Fatalf("Error parsing guild file: %v", err)
	}
	fmt.Printf("Processed %d players from guild.txt\n", len(guildPlayers))

	// Count online players
	onlineCount := 0
	for _, player := range guildPlayers {
		if player.Status == "Online" {
			onlineCount++
		}
	}
	fmt.Printf("Found %d online players in guild\n", onlineCount)

	// Parse sheet file
	fmt.Println("Reading sheet data...")
	sheetNames, err := parseSheetFile("data/sheet.txt")
	if err != nil {
		log.Fatalf("Error parsing sheet file: %v", err)
	}
	fmt.Printf("Processed %d player names from sheet.txt\n", len(sheetNames))

	// Find players online but not in sheet
	fmt.Println("Analyzing data...")
	missingPlayers, excludedPlayers, guildMatches := findOnlinePlayersNotInSheet(guildPlayers, sheetNames, altNames)

	// Find players in sheet but not in guild
	sheetPlayersNotInGuild, sheetMatches := findSheetPlayersNotInGuild(guildPlayers, sheetNames, altNames)

	// Show successful matches first
	if len(guildMatches) > 0 {
		fmt.Printf("\n=== SUCCESSFUL MATCHES ===\n")
		directMatches := 0
		alternativeMatches := 0
		ignoredMatches := 0

		for _, match := range guildMatches {
			switch match.MatchType {
			case "direct":
				directMatches++
			case "alternative":
				fmt.Printf("Matched: %s (found as '%s' in sheet)\n", match.GuildName, match.AlternativeName)
				alternativeMatches++
			case "ignored":
				fmt.Printf("Matched: %s (pattern match with '%s' in sheet)\n", match.GuildName, match.AlternativeName)
				ignoredMatches++
			}
		}

		fmt.Printf("- Direct matches: %d\n", directMatches)
		fmt.Printf("- Alternative name matches: %d\n", alternativeMatches)
		if ignoredMatches > 0 {
			fmt.Printf("- Pattern matches: %d\n", ignoredMatches)
		}
	}

	// Output results
	fmt.Printf("\n=== RESULTS ===\n")
	fmt.Printf("Players online but not in sheet (%d):\n", len(missingPlayers))

	if len(missingPlayers) == 0 {
		fmt.Println("  (none)")
	} else {
		for i, player := range missingPlayers {
			if i == len(missingPlayers)-1 {
				fmt.Printf("  %s\n", player)
			} else {
				fmt.Printf("  %s,\n", player)
			}
		}
	}

	// Show excluded players
	if len(excludedPlayers) > 0 {
		fmt.Printf("\nExcluded players (have special roles) (%d):\n", len(excludedPlayers))
		for i, player := range excludedPlayers {
			if i == len(excludedPlayers)-1 {
				fmt.Printf("  %s\n", player)
			} else {
				fmt.Printf("  %s,\n", player)
			}
		}
	}

	// Show players in sheet but not in guild
	if len(sheetPlayersNotInGuild) > 0 {
		fmt.Printf("\nPlayers in sheet but not in guild (%d):\n", len(sheetPlayersNotInGuild))
		for i, player := range sheetPlayersNotInGuild {
			if i == len(sheetPlayersNotInGuild)-1 {
				fmt.Printf("  %s\n", player)
			} else {
				fmt.Printf("  %s,\n", player)
			}
		}
	}

	// Show sheet matches if any
	/*
		if len(sheetMatches) > 0 {
			fmt.Printf("\nSheet name matches found (%d):\n", len(sheetMatches))
			for _, match := range sheetMatches {
				if match.MatchType == "alternative" {
					fmt.Printf("  '%s' in sheet -> %s in guild\n", match.AlternativeName, match.GuildName)
				}
			}
		}
	*/

	fmt.Printf("\nSummary:\n")
	fmt.Printf("- Total guild members: %d\n", len(guildPlayers))
	fmt.Printf("- Online guild members: %d\n", onlineCount)
	fmt.Printf("- Players in sheet: %d\n", len(sheetNames))
	fmt.Printf("- Successful matches: %d\n", len(guildMatches)+len(sheetMatches))
	fmt.Printf("- Online players missing from sheet: %d\n", len(missingPlayers))
	fmt.Printf("- Excluded players (special roles): %d\n", len(excludedPlayers))
	fmt.Printf("- Sheet players not in guild: %d\n", len(sheetPlayersNotInGuild))
}
