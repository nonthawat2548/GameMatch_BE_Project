-- Fictional fixtures, clearly labelled. RAWG import supplies actual game metadata.
INSERT INTO gm_games(title,description,genres,tags,platforms)
SELECT x.title,x.description,x.genres,x.tags,ARRAY['PC']::text[] FROM (VALUES
 ('Demo: Story Quest','Fictional RPG fixture for testing the matching engine.',ARRAY['rpg'],ARRAY['singleplayer','story-rich']),
 ('Demo: Team Arena','Fictional PvP fixture.',ARRAY['action'],ARRAY['pvp','competitive']),
 ('Demo: Cozy Garden','Fictional relaxing co-op fixture.',ARRAY['simulation'],ARRAY['co-op','relaxing']),
 ('Demo: Hard Expedition','Fictional challenging adventure fixture.',ARRAY['adventure'],ARRAY['singleplayer','difficult'])
) AS x(title,description,genres,tags)
WHERE NOT EXISTS(SELECT 1 FROM gm_games g WHERE g.title=x.title);
