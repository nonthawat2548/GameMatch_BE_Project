-- Preserve existing rows. Any conflicting data aborts this entire statement.
-- Supabase uses public; isolated tests set their own search_path.
DO $$
BEGIN
 PERFORM set_config('lock_timeout','5s',true);
 IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='gm_members'::regclass AND conname='gm_members_username_key') THEN
  ALTER TABLE gm_members ADD CONSTRAINT gm_members_username_key UNIQUE (username);
 END IF;
 IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='gm_members'::regclass AND conname='gm_members_username_identity_check') THEN
  ALTER TABLE gm_members ADD CONSTRAINT gm_members_username_identity_check
   CHECK (position('@' in username)=0 OR username=email);
 END IF;
 IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='gm_members'::regclass AND conname='gm_members_display_name_check') THEN
  ALTER TABLE gm_members ADD CONSTRAINT gm_members_display_name_check
   CHECK (btrim(display_name)<>'' AND char_length(display_name) BETWEEN 1 AND 50);
 END IF;
 IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='gm_reviews'::regclass AND conname='gm_reviews_comment_check') THEN
  ALTER TABLE gm_reviews ADD CONSTRAINT gm_reviews_comment_check
   CHECK (char_length(btrim(comment))>=10 AND octet_length(comment)<=2000);
 END IF;
END $$;
