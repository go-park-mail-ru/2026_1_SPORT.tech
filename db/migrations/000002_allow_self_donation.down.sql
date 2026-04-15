ALTER TABLE donation
ADD CONSTRAINT donation_check
CHECK (sender_user_id <> recipient_user_id);
