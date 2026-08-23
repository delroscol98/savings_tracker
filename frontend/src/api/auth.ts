import {
  LoginResponseSchema,
  MessageSchema,
  UserSchema,
  type LoginResponse,
  type Message,
  type User,
} from "./schemas";
import { client } from "./client";

export async function register(
  email: string,
  password: string,
  full_name: string,
): Promise<User> {
  return client<User>(
    "/api/users",
    {
      method: "POST",
      body: { email, password, full_name },
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

export async function resetPassword(
  token: string,
  password: string,
): Promise<Message> {
  return client<Message>(
    "/api/reset-password",
    {
      method: "POST",
      body: {
        token,
        password,
      },
    },
    MessageSchema,
  );
}
