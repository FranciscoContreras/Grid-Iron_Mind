-- Fix team conference/division/stadium metadata
-- Run this with: heroku pg:psql -a grid-iron-mind < scripts/fix_team_metadata.sql

-- AFC East
UPDATE teams SET conference = 'AFC', division = 'East', stadium = 'Highmark Stadium' WHERE abbreviation = 'BUF';
UPDATE teams SET conference = 'AFC', division = 'East', stadium = 'Hard Rock Stadium' WHERE abbreviation = 'MIA';
UPDATE teams SET conference = 'AFC', division = 'East', stadium = 'Gillette Stadium' WHERE abbreviation = 'NE';
UPDATE teams SET conference = 'AFC', division = 'East', stadium = 'MetLife Stadium' WHERE abbreviation = 'NYJ';

-- AFC North
UPDATE teams SET conference = 'AFC', division = 'North', stadium = 'M&T Bank Stadium' WHERE abbreviation = 'BAL';
UPDATE teams SET conference = 'AFC', division = 'North', stadium = 'Paycor Stadium' WHERE abbreviation = 'CIN';
UPDATE teams SET conference = 'AFC', division = 'North', stadium = 'Cleveland Browns Stadium' WHERE abbreviation = 'CLE';
UPDATE teams SET conference = 'AFC', division = 'North', stadium = 'Acrisure Stadium' WHERE abbreviation = 'PIT';

-- AFC South
UPDATE teams SET conference = 'AFC', division = 'South', stadium = 'NRG Stadium' WHERE abbreviation = 'HOU';
UPDATE teams SET conference = 'AFC', division = 'South', stadium = 'Lucas Oil Stadium' WHERE abbreviation = 'IND';
UPDATE teams SET conference = 'AFC', division = 'South', stadium = 'EverBank Stadium' WHERE abbreviation = 'JAX';
UPDATE teams SET conference = 'AFC', division = 'South', stadium = 'Nissan Stadium' WHERE abbreviation = 'TEN';

-- AFC West
UPDATE teams SET conference = 'AFC', division = 'West', stadium = 'Empower Field at Mile High' WHERE abbreviation = 'DEN';
UPDATE teams SET conference = 'AFC', division = 'West', stadium = 'GEHA Field at Arrowhead Stadium' WHERE abbreviation = 'KC';
UPDATE teams SET conference = 'AFC', division = 'West', stadium = 'Allegiant Stadium' WHERE abbreviation = 'LV';
UPDATE teams SET conference = 'AFC', division = 'West', stadium = 'SoFi Stadium' WHERE abbreviation = 'LAC';

-- NFC East
UPDATE teams SET conference = 'NFC', division = 'East', stadium = 'AT&T Stadium' WHERE abbreviation = 'DAL';
UPDATE teams SET conference = 'NFC', division = 'East', stadium = 'MetLife Stadium' WHERE abbreviation = 'NYG';
UPDATE teams SET conference = 'NFC', division = 'East', stadium = 'Lincoln Financial Field' WHERE abbreviation = 'PHI';
UPDATE teams SET conference = 'NFC', division = 'East', stadium = 'Northwest Stadium' WHERE abbreviation = 'WSH';

-- NFC North
UPDATE teams SET conference = 'NFC', division = 'North', stadium = 'Soldier Field' WHERE abbreviation = 'CHI';
UPDATE teams SET conference = 'NFC', division = 'North', stadium = 'Ford Field' WHERE abbreviation = 'DET';
UPDATE teams SET conference = 'NFC', division = 'North', stadium = 'Lambeau Field' WHERE abbreviation = 'GB';
UPDATE teams SET conference = 'NFC', division = 'North', stadium = 'U.S. Bank Stadium' WHERE abbreviation = 'MIN';

-- NFC South
UPDATE teams SET conference = 'NFC', division = 'South', stadium = 'Mercedes-Benz Stadium' WHERE abbreviation = 'ATL';
UPDATE teams SET conference = 'NFC', division = 'South', stadium = 'Bank of America Stadium' WHERE abbreviation = 'CAR';
UPDATE teams SET conference = 'NFC', division = 'South', stadium = 'Caesars Superdome' WHERE abbreviation = 'NO';
UPDATE teams SET conference = 'NFC', division = 'South', stadium = 'Raymond James Stadium' WHERE abbreviation = 'TB';

-- NFC West
UPDATE teams SET conference = 'NFC', division = 'West', stadium = 'State Farm Stadium' WHERE abbreviation = 'ARI';
UPDATE teams SET conference = 'NFC', division = 'West', stadium = 'Levi''s Stadium' WHERE abbreviation = 'SF';
UPDATE teams SET conference = 'NFC', division = 'West', stadium = 'Lumen Field' WHERE abbreviation = 'SEA';
UPDATE teams SET conference = 'NFC', division = 'West', stadium = 'SoFi Stadium' WHERE abbreviation = 'LAR';
