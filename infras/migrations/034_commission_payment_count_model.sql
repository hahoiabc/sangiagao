-- Switch commission stages from time-based to payment-count-based.
--
-- OLD: stage 1 (0-90d) 50% / stage 2 (91-270d) 30% / stage 3 (271+d) 20%
-- NEW: payment_sequence 1 → 45%, payment 2 → 30%, payment 3+ → 15%
--
-- Rationale: simpler ("lần 1, lần 2, lần 3+"), pushes annual plans (aff
-- earns more per long-cycle sub), behaviour-fair (monthly subscribers no
-- longer get inflated commission from staying in stage 1 forever).
--
-- Migration strategy B: existing records keep their frozen stage/rate;
-- only NEW commission records (created from now on) use the new rates.

-- 1. Update default rule rates. stage_1_days / stage_2_days columns are
-- repurposed as "payment count threshold" but the engine ignores them now
-- (uses payment_sequence directly). Kept for backward-compat schema.
UPDATE commission_rules
   SET stage1_pct = 0.45,
       stage2_pct = 0.30,
       stage3_pct = 0.15,
       stage1_days = 1,     -- semantic: payment #1 = stage 1
       stage2_days = 1,     -- semantic: payment #2 = stage 2 (then 3+ stage 3)
       updated_at = NOW()
 WHERE referral_code_id IS NULL;  -- default rule only

-- 2. Add payment_sequence to commission_records for analytics/audit.
-- Existing records default to 1 (most are first payment anyway; analytics
-- can re-derive precise sequence later via window function if needed).
ALTER TABLE commission_records
  ADD COLUMN IF NOT EXISTS payment_sequence INT NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_commission_records_ref_pair
  ON commission_records(referrer_user_id, referee_user_id);
