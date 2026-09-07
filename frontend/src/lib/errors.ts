import { ApiError } from "@/api/client";
import { toast } from "sonner";

export function handleApiError(
  error: unknown,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  setError: any,
  toastId?: string | number,
) {
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

  setError("root", { message: apiError.error });
  toast.error(apiError.error, { id: toastId });
  if (apiError.fields) {
    for (const [field, messages] of Object.entries(apiError.fields)) {
      setError(field, { message: messages[0] });
    }
  }
}
