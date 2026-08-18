import { z } from "zod";
import { LoginResponseSchema, MessageSchema, UserSchema } from "./schemas";
import { client } from "./client";

type User = z.infer<typeof UserSchema>;
type LoginResponse = z.infer<typeof LoginResponseSchema>;
type Message = z.infer<typeof MessageSchema>;

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

export async function forgotPassword(email: string): Promise<Message> {
  return client<Message>(
    "/api/forgot-password",
    {
      method: "POST",
      body: { email },
    },
    MessageSchema,
  );
}

export async function resetPassword(password: string): Promise<Message> {
  return client<Message>(
    "/api/reset-password",
    {
      method: "POST",
      body: {
        password,
      },
    },
    MessageSchema,
  );
}
