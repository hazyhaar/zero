package vtparser

// interval représente une plage inclusive [first, last].
type interval struct {
	first rune
	last  rune
}

// bsearchInterval teste si r est présent dans la table d'intervalles ordonnée.
func bsearchInterval(r rune, table []interval) bool {
	min := 0
	max := len(table) - 1
	for max >= min {
		mid := (min + max) / 2
		if r > table[mid].last {
			min = mid + 1
		} else if r < table[mid].first {
			max = mid - 1
		} else {
			return true
		}
	}
	return false
}

// zeroWidthIntervals contient les points de code Unicode de largeur 0 (marques combinatoires, espaces nuls).
var zeroWidthIntervals = []interval{
	{0x0300, 0x036F},   // Combining Diacritical Marks
	{0x0483, 0x0489},   // Cyrillic Combining Marks
	{0x0591, 0x05BD},   // Hebrew Cantillation Marks
	{0x05BF, 0x05BF},
	{0x05C1, 0x05C2},
	{0x05C4, 0x05C5},
	{0x05C7, 0x05C7},
	{0x0610, 0x061A},   // Arabic Combining Marks
	{0x064B, 0x065F},
	{0x0670, 0x0670},
	{0x06D6, 0x06DC},
	{0x06DF, 0x06E4},
	{0x06E7, 0x06E8},
	{0x06EA, 0x06ED},
	{0x0711, 0x0711},
	{0x0730, 0x074A},
	{0x07A6, 0x07B0},
	{0x07EB, 0x07F3},
	{0x0901, 0x0902},   // Devanagari Marks
	{0x093C, 0x093C},
	{0x0941, 0x0948},
	{0x094D, 0x094D},
	{0x0951, 0x0957},
	{0x0962, 0x0963},
	{0x0E31, 0x0E31},   // Thai Marks
	{0x0E34, 0x0E3A},
	{0x0E47, 0x0E4E},
	{0x1AB0, 0x1AFF},   // Combining Diacritical Marks Extended
	{0x1DC0, 0x1DFF},   // Combining Diacritical Marks Supplement
	{0x200B, 0x200F},   // Zero Width Space / Format Controls
	{0x202A, 0x202E},   // BiDi Controls
	{0x2060, 0x206F},   // Invisible Operators
	{0x20D0, 0x20FF},   // Combining Marks for Symbols
	{0xFE00, 0xFE0F},   // Variation Selectors
	{0xFE20, 0xFE2F},   // Combining Half Marks
	{0xFEFF, 0xFEFF},   // Zero Width No-Break Space (BOM)
	{0xE0100, 0xE01EF}, // Variation Selectors Supplement
}

// wideIntervals contient les plages Unicode East Asian Wide / Fullwidth et Émojis (largeur 2).
var wideIntervals = []interval{
	{0x1100, 0x115F},   // Hangul Jamo
	{0x231A, 0x231B},   // Watch, Hourglass
	{0x2329, 0x232A},   // Left/Right-Pointing Angle Bracket
	{0x23E9, 0x23EC},   // Fast Forward, Fast Rewind, etc.
	{0x23F0, 0x23F0},   // Alarm Clock
	{0x23F3, 0x23F3},   // Hourglass with Flowing Sand
	{0x25FD, 0x25FE},   // Medium Black/White Squares
	{0x2614, 0x2615},   // Umbrella with Rain Drops, Hot Beverage
	{0x2648, 0x2653},   // Zodiac Symbols
	{0x267F, 0x267F},   // Wheelchair Symbol
	{0x2693, 0x2693},   // Anchor
	{0x26A1, 0x26A1},   // High Voltage
	{0x26AA, 0x26AB},   // White/Black Medium Circle
	{0x26BD, 0x26BE},   // Soccer Ball, Baseball
	{0x26C4, 0x26C5},   // Snowman Without Snow, Sun Behind Cloud
	{0x26CE, 0x26CE},   // Ophiuchus
	{0x26D4, 0x26D4},   // No Entry
	{0x26EA, 0x26EA},   // Church
	{0x26F2, 0x26F3},   // Fountain, Flag in Hole
	{0x26F5, 0x26F5},   // Sailboat
	{0x26FA, 0x26FA},   // Tent
	{0x26FD, 0x26FD},   // Fuel Pump
	{0x2705, 0x2705},   // White Heavy Check Mark
	{0x270A, 0x270B},   // Raised Fist, Raised Hand
	{0x2728, 0x2728},   // Sparkles
	{0x274C, 0x274C},   // Cross Mark
	{0x274E, 0x274E},   // Negative Squared Cross Mark
	{0x2753, 0x2755},   // Question Marks
	{0x2757, 0x2757},   // Heavy Exclamation Mark Symbol
	{0x2795, 0x2797},   // Heavy Plus, Minus, Division
	{0x27B0, 0x27B0},   // Curly Loop
	{0x27BF, 0x27BF},   // Double Curly Loop
	{0x2B1B, 0x2B1C},   // Black/White Large Square
	{0x2B50, 0x2B50},   // White Medium Star
	{0x2B55, 0x2B55},   // Heavy Large Circle
	{0x2E80, 0x2FFF},   // CJK Radicals / Kangxi Radicals / Ideographic Description
	{0x3000, 0x303E},   // CJK Symbols and Punctuation (Id. Space U+3000)
	{0x3040, 0x309F},   // Hiragana
	{0x30A0, 0x30FF},   // Katakana
	{0x3100, 0x312F},   // Bopomofo
	{0x3130, 0x318F},   // Hangul Compatibility Jamo
	{0x3190, 0x31E3},   // Kanbun / CJK Strokes
	{0x3200, 0x32FF},   // Enclosed CJK Letters and Months
	{0x3300, 0x33FF},   // CJK Compatibility
	{0x3400, 0x4DBF},   // CJK Unified Ideographs Extension A
	{0x4E00, 0x9FFF},   // CJK Unified Ideographs
	{0xA960, 0xA97F},   // Hangul Jamo Extended-A
	{0xAC00, 0xD7A3},   // Hangul Syllables
	{0xD7B0, 0xD7FF},   // Hangul Jamo Extended-B
	{0xF900, 0xFAFF},   // CJK Compatibility Ideographs
	{0xFE10, 0xFE19},   // Vertical Forms
	{0xFE30, 0xFE6F},   // CJK Compatibility Forms / Small Form Variants
	{0xFF01, 0xFF60},   // Fullwidth Forms (ASCII fullwidth)
	{0xFFE0, 0xFFE6},   // Fullwidth Symbol Variants
	{0x16FE0, 0x16FFF}, // Ideographic Symbols and Punctuation
	{0x17000, 0x187FF}, // Tangut
	{0x18800, 0x18AFF}, // Tangut Components
	{0x1B000, 0x1B0FF}, // Kana Supplement
	{0x1F000, 0x1F02F}, // Mahjong Tiles
	{0x1F0A0, 0x1F0FF}, // Playing Cards
	{0x1F100, 0x1F64F}, // Enclosed Alphanumeric Supp, Misc Symbols & Pictographs, Emoticons
	{0x1F680, 0x1F6FF}, // Transport and Map Symbols
	{0x1F700, 0x1F77F}, // Alchemical Symbols
	{0x1F780, 0x1F7FF}, // Geometric Shapes Extended
	{0x1F800, 0x1F8FF}, // Supplemental Arrows-C
	{0x1F900, 0x1F9FF}, // Supplemental Symbols and Pictographs
	{0x1FA00, 0x1FAFF}, // Symbols and Pictographs Extended-A
	{0x20000, 0x2FFFD}, // CJK Unified Ideographs Extension B..F
	{0x30000, 0x3FFFD}, // CJK Unified Ideographs Extension G..
}

// RuneWidth calcule la largeur d'affichage en colonnes d'un point de code Unicode.
// - 0 : Caractères de contrôle, marques combinatoires et formatage nul.
// - 1 : ASCII standard et caractères occidentaux.
// - 2 : Idéogrammes CJK, émojis et formes pleine largeur.
func RuneWidth(r rune) int {
	// Chemin rapide ASCII standard
	if r >= 0x20 && r < 0x7F {
		return 1
	}
	if r < 0x20 || (r >= 0x7F && r < 0xA0) {
		return 0
	}

	// Recherche dans les tables d'intervalles
	if bsearchInterval(r, zeroWidthIntervals) {
		return 0
	}
	if bsearchInterval(r, wideIntervals) {
		return 2
	}

	return 1
}
