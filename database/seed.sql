-- Fictional fixtures, clearly labelled. RAWG import supplies actual game metadata.
INSERT INTO gm_games(title,description,genres,tags,platforms)
SELECT x.title,x.description,x.genres,x.tags,ARRAY['PC']::text[] FROM (VALUES
 ('Demo: Story Quest','Fictional RPG fixture for testing the matching engine.',ARRAY['rpg'],ARRAY['singleplayer','story-rich']),
 ('Demo: Team Arena','Fictional PvP fixture.',ARRAY['action'],ARRAY['pvp','competitive']),
 ('Demo: Cozy Garden','Fictional relaxing co-op fixture.',ARRAY['simulation'],ARRAY['co-op','relaxing']),
 ('Demo: Hard Expedition','Fictional challenging adventure fixture.',ARRAY['adventure'],ARRAY['singleplayer','difficult'])
) AS x(title,description,genres,tags)
WHERE NOT EXISTS(SELECT 1 FROM gm_games g WHERE g.title=x.title);

INSERT INTO gm_genres(slug)
SELECT DISTINCT genre FROM gm_games CROSS JOIN LATERAL unnest(genres) AS x(genre)
WHERE genre IS NOT NULL AND btrim(genre) <> '' ORDER BY genre
ON CONFLICT (slug) DO NOTHING;
INSERT INTO gm_game_genres(game_id,genre_id,position)
SELECT g.id,d.id,min(x.pos)::integer
FROM gm_games g CROSS JOIN LATERAL unnest(g.genres) WITH ORDINALITY AS x(genre,pos)
JOIN gm_genres d ON d.slug=x.genre
WHERE NOT EXISTS (SELECT 1 FROM gm_game_genres existing WHERE existing.game_id=g.id)
GROUP BY g.id,d.id ON CONFLICT (game_id,genre_id) DO NOTHING;
