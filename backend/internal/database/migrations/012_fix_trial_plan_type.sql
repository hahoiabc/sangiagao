-- Fix trial subscriptions incorrectly stored as 'paid'
-- Trial subscriptions have amount = 0, should be plan = 'free_trial'
UPDATE subscriptions SET plan = 'free_trial' WHERE amount = 0 AND plan = 'paid';
