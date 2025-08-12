# Signup Checker

A Go script that identifies guild members who are online but not listed in the signup sheet, with advanced name matching capabilities.

## Features

- **Alternative Name Mapping**: Maps guild names to alternative names used in signup sheets
- **Enhanced Logging**: Shows exactly how names were matched (direct, alternative, or pattern)
- **Role-based Exclusions**: Automatically excludes players with special roles (Bombers, Guild Master)
- **Bidirectional Analysis**: Finds both missing players and players in sheet but not in guild
- **Clean Output**: Detailed results with match statistics

## File Structure

```
data/
├── guild.txt          # Guild member data (tab-separated, quoted fields)
├── sheet.txt          # Signup sheet names (one per line)
└── sheet-names.txt    # Alternative name mappings
```

## Alternative Names File Format

`data/sheet-names.txt` maps guild names to alternative names:

```
# Format: GuildName:AlternativeName1,AlternativeName2
xSarge:Sarge,sarge
Boneappletea:boner,bone
ClapYourFace:Clap,clap
Jbeil:JB,jb
```

## Usage

```bash
# Run with Go
go run main.go

# Or use the compiled executable
./signup-checker.exe
```

## Output

The script provides:

1. **Alternative name mappings loaded**
2. **Successful matches section** showing:
   - Direct matches (exact name matches)
   - Alternative name matches with details
   - Pattern matches (legacy support)
3. **Results section** showing:
   - Players online but not in sheet
   - Excluded players (special roles)
   - Players in sheet but not in guild
4. **Summary statistics**

## Example Output

```
=== SUCCESSFUL MATCHES ===
Matched: xSarge (found as 'Sarge' in sheet)
Matched: Boneappletea (found as 'boner' in sheet)
- Direct matches: 34
- Alternative name matches: 13

=== RESULTS ===
Players online but not in sheet (4):
  cheholahola, Hiccup, NightStrikerBigT, NordtonSP

Summary:
- Successful matches: 97
- Online players missing from sheet: 4
```

## Role Exclusions

Players with these roles are automatically excluded:
- **Bomber** - Special combat role
- **Guild Master** - Guild leader

## Requirements

- Go 1.21 or later
- Input files in `data/` directory
