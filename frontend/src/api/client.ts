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
  options: RequestInit & { body?: unknown },
): Promise<T> {
  const url = import.meta.env.VITE_API_URL + path;
  const headers: Record<string, string> = {};
  const token = getStoredToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  if (options.body && typeof options.body != "string") {
    options.body = JSON.stringify(options.body);
    headers["Content-Type"] = "application/json";
  }

  let response: Response;
  try {
    response = await fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        ...headers,
      },
    });
  } catch {
    throw new ApiError(0, "Network error");
  }

  let data: any;
  try {
    data = await response.json();
  } catch {
    throw new ApiError(response.status, "Invalid response");
  }

  if (!response.ok) {
    throw new ApiError(response.status, data.error, data.fields);
  }

  return data as T;
}
