import { z } from "zod";

export const UserSchema = z.object({
  id: z.uuid(),
  created_at: z.coerce.date(),
  updated_at: z.coerce.date(),
  email: z.email(),
});

export const LoginResponseSchema = z
  .object({
    token: z.string(),
  })
  .extend(UserSchema);

export const GoalSchema = z.object({
  id: z.uuid(),
  target: z.number().int().nonnegative(),
  deadline: z.coerce.date(),
  user_id: z.uuid(),
  progress: z.number().int(),
});

export const DepositSchema = z.object({
  id: z.uuid(),
  amount: z.number().int().positive(),
  note: z.string().optional(),
  created_at: z.coerce.date(),
});

export const ErrorEnvelope = z.object({
  error: z.string(),
  fields: z.record(z.string(), z.array(z.string())).optional(),
});
