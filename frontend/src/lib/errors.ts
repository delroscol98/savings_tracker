import { ApiError } from "@/api/client";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function handleApiError(error: unknown, setError: any) {
  const apiError =
    error instanceof Error && error.cause instanceof ApiError
      ? error.cause
      : error instanceof ApiError
        ? error
        : null;

  if (!apiError) {
    setError("root", { message: "An unexpected error occurred" });
    return;
  }

  if (apiError.fields) {
    for (const [field, messages] of Object.entries(apiError.fields)) {
      setError(field, { message: messages[0] });
    }
  } else {
    setError("root", { message: apiError.error });
  }
}
