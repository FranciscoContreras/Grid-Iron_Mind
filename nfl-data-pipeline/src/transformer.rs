use std::collections::HashMap;
use once_cell::sync::Lazy;

/// Map historical team abbreviations to current ones
static TEAM_MAPPING: Lazy<HashMap<&'static str, &'static str>> = Lazy::new(|| {
    let mut m = HashMap::new();

    // Historical team name changes
    m.insert("STL", "LA");    // St. Louis Rams → Los Angeles Rams (2016)
    m.insert("SD", "LAC");    // San Diego Chargers → Los Angeles Chargers (2017)
    m.insert("OAK", "LV");    // Oakland Raiders → Las Vegas Raiders (2020)

    // Legacy abbreviations
    m.insert("SL", "LA");     // Alternative St. Louis abbreviation
    m.insert("BLT", "BAL");   // Baltimore (legacy)
    m.insert("CLV", "CLE");   // Cleveland (legacy)
    m.insert("HST", "HOU");   // Houston (legacy)
    m.insert("ARZ", "ARI");   // Arizona (legacy)

    m
});

pub fn normalize_team_abbr(abbr: &str) -> String {
    TEAM_MAPPING
        .get(abbr)
        .map(|s| s.to_string())
        .unwrap_or_else(|| abbr.to_uppercase())
}

/// Convert height string (e.g., "6-2") to inches
pub fn height_to_inches(height_str: &str) -> Option<i32> {
    let parts: Vec<&str> = height_str.split('-').collect();
    if parts.len() == 2 {
        let feet = parts[0].parse::<i32>().ok()?;
        let inches = parts[1].parse::<i32>().ok()?;
        Some(feet * 12 + inches)
    } else {
        None
    }
}

/// Normalize player position
pub fn normalize_position(pos: &str) -> String {
    match pos.to_uppercase().as_str() {
        "HB" => "RB".to_string(),
        "FB" => "RB".to_string(),
        "ILB" | "OLB" | "MLB" => "LB".to_string(),
        "CB" | "S" | "FS" | "SS" => "DB".to_string(),
        "DE" | "DT" | "NT" => "DL".to_string(),
        p => p.to_string(),
    }
}

/// Clean player name (remove suffixes like Jr., III, etc.)
pub fn clean_player_name(name: &str) -> String {
    name.replace(" Jr.", "")
        .replace(" Sr.", "")
        .replace(" II", "")
        .replace(" III", "")
        .replace(" IV", "")
        .trim()
        .to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_team_mapping() {
        assert_eq!(normalize_team_abbr("STL"), "LA");
        assert_eq!(normalize_team_abbr("SD"), "LAC");
        assert_eq!(normalize_team_abbr("OAK"), "LV");
        assert_eq!(normalize_team_abbr("KC"), "KC");
    }

    #[test]
    fn test_height_conversion() {
        assert_eq!(height_to_inches("6-2"), Some(74));
        assert_eq!(height_to_inches("5-11"), Some(71));
        assert_eq!(height_to_inches("invalid"), None);
    }

    #[test]
    fn test_position_normalization() {
        assert_eq!(normalize_position("HB"), "RB");
        assert_eq!(normalize_position("ILB"), "LB");
        assert_eq!(normalize_position("QB"), "QB");
    }
}
