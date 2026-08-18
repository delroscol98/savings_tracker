import { type ZodType } from "zod";
import { ErrorEnvelope } from "./schemas";

export function getStoredToken(): string {
  const token = localStorage.getItem("token");
  if (token == undefined) {
    return "";
  }

  return token;
}

export function setStoredToken(token: string): void {
  localStorage.setItem("token", token);
}

export function clearStoredToken(): void {
  localStorage.removeItem("token");
}

export class ApiError extends Error {
  status: number;
  error: string;
  fields?: Record<string, string[]>;

  constructor(
    status: number,
    error: string,
    fields?: Record<string, string[]>,
  ) {
    super(error);
    this.name = "ApiError";
    this.status = status;
    this.error = error;
    this.fields = fields;
  }
}

export async function client<T>(
  path: string,
  options: Omit<RequestInit, "body"> & { body?: unknown },
  schema?: ZodType<T>,
): Promise<T> {
  const url = import.meta.env.VITE_API_URL + path;
  const headers: Record<string, string> = {};
  const token = getStoredToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const { body, ...rest } = options;
  if (body && typeof body != "string") {
    options.body = JSON.stringify(body);
    headers["Content-Type"] = "application/json";
  }

  let response: Response;
  try {
    response = await fetch(url, {
      ...rest,
      body: typeof body === "string" ? body : undefined,
      headers: { ...rest.headers, ...headers },
    });
  } catch {
    throw new ApiError(0, "Network error");
  }

  let data: unknown;
  try {
    data = await response.json();
  } catch {
    throw new ApiError(response.status, "Invalid response");
  }

  if (!response.ok) {
    const result = ErrorEnvelope.safeParse(data);
    if (result.success) {
      throw new ApiError(
        response.status,
        result.data.error,
        result.data.fields,
      );
    }
    throw new ApiError(response.status, "Unknown error type");
  }

  if (schema) {
    const result = schema.safeParse(data);
    if (!result.success) {
      throw new ApiError(
        response.status,
        `Response validation failed: ${result.error.message}`,
      );
    }
    return result.data;
  }

  return data as T;
}
