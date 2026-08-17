import { z } from "zod";
import {
  ForgotPasswordSchema,
  LoginResponseSchema,
  ResetPasswordSchema,
  UserSchema,
} from "./schemas";
import { client } from "./client";

type User = z.infer<typeof UserSchema>;
type LoginResponse = z.infer<typeof LoginResponseSchema>;
type ForgotPassword = z.infer<typeof ForgotPasswordSchema>;
type ResetPassword = z.infer<typeof ResetPasswordSchema>;

export async function register(
  email: string,
  password: string,
  fullName: string,
): Promise<User> {
  return client<User>(
    "/api/users",
    {
      method: "POST",
      body: { email, password, fullName },
    },
    UserSchema,
  );
}

export async function login(
  email: string,
  password: string,
): Promise<LoginResponse> {
  return client<LoginResponse>(
    "/api/login",
    {
      method: "POST",
      body: { email, password },
    },
    LoginResponseSchema,
  );
}

export async function forgotPassword(email: string): Promise<ForgotPassword> {
  return client<ForgotPassword>(
    "/api/forgot-password",
    {
      method: "POST",
      body: { email },
    },
    ForgotPasswordSchema,
  );
}

export async function resetPassword(password: string): Promise<ResetPassword> {
  return client<ResetPassword>(
    "/api/reset-password",
    {
      method: "POST",
      body: {
        password,
      },
    },
    ResetPasswordSchema,
  );
}
