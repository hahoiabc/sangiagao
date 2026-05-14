-- Add list_amount (giá niêm yết) to subscription_plans
-- Frontend displays list_amount struck-through next to amount when > 0.
ALTER TABLE subscription_plans
  ADD COLUMN IF NOT EXISTS list_amount BIGINT NOT NULL DEFAULT 0;

-- Seed list price = giá IAP (Apple/Google) cho 4 gói mặc định.
UPDATE subscription_plans SET list_amount = 49000  WHERE months = 1  AND list_amount = 0;
UPDATE subscription_plans SET list_amount = 147000 WHERE months = 3  AND list_amount = 0;
UPDATE subscription_plans SET list_amount = 294000 WHERE months = 6  AND list_amount = 0;
UPDATE subscription_plans SET list_amount = 588000 WHERE months = 12 AND list_amount = 0;
