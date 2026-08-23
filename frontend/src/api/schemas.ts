import { z } from "zod";

export const UserSchema = z.object({
  id: z.uuid(),
  created_at: z.coerce.date(),
  updated_at: z.coerce.date(),
  email: z.email(),
});
export type User = z.infer<typeof UserSchema>;

export const LoginResponseSchema = UserSchema.extend({
  token: z.string(),
});
export type LoginResponse = z.infer<typeof LoginResponseSchema>;

export const MessageSchema = z.object({
  message: z.string(),
});
export type Message = z.infer<typeof MessageSchema>;

export const GoalSchema = z.object({
  id: z.uuid(),
  target: z.number().int().nonnegative(),
  deadline: z.coerce.date(),
  user_id: z.uuid(),
  progress: z.number().int(),
});
export type Goal = z.infer<typeof GoalSchema>;

export const DepositSchema = z.object({
  id: z.uuid(),
  amount: z.number().int().positive(),
  note: z.string().optional(),
  created_at: z.coerce.date(),
});
export type Deposit = z.infer<typeof DepositSchema>;

export const ErrorEnvelopeSchema = z.object({
  error: z.string(),
  fields: z.record(z.string(), z.array(z.string())).optional(),
});
export type ErrorEnvelope = z.infer<typeof ErrorEnvelopeSchema>;
