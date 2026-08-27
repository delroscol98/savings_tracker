import { useSearchParams, Link, useNavigate } from "react-router";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { resetPassword } from "@/api/auth";
import { handleApiError } from "@/lib/errors";

const resetPasswordSchema = z.object({
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(128, "Password must be at most 128 characters"),
});

type ResetPasswordForm = z.infer<typeof resetPasswordSchema>;

export function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<ResetPasswordForm>({
    resolver: zodResolver(resetPasswordSchema),
  });

  if (!token) {
    return (
      <div className="min-h-screen bg-neutral-900 flex items-center justify-center px-spacing-7">
        <div className="bg-neutral-800 rounded-radius-2xl p-spacing-12 w-full max-w-[400px]">
          <h1 className="text-3 text-neutral-0 mb-spacing-7">
            Invalid reset link
          </h1>
          <p className="text-5 text-neutral-300 mb-spacing-9">
            This password reset link is invalid or has expired.
          </p>
          <Link
            to="/forgot-password"
            className="block text-center text-6 text-orange-500 hover:text-orange-400"
          >
            Request a new reset link
          </Link>
        </div>
      </div>
    );
  }

  async function onSubmit(data: ResetPasswordForm) {
    try {
      await resetPassword(token!, data.password);
      navigate("/login");
    } catch (error: unknown) {
      handleApiError(error, setError);
    }
  }

  return (
    <div className="min-h-screen bg-neutral-900 flex items-center justify-center px-spacing-7">
      <div className="bg-neutral-800 rounded-radius-2xl p-spacing-12 w-full max-w-[400px]">
        <h1 className="text-3 text-neutral-0 mb-spacing-10">
          Reset password
        </h1>

        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-spacing-7">
          {errors.root && (
            <p className="text-red-500 text-6">{errors.root.message}</p>
          )}

          <div className="flex flex-col gap-spacing-3">
            <label htmlFor="password" className="text-6 text-neutral-300">
              New password
            </label>
            <input
              id="password"
              type="password"
              {...register("password")}
              className="bg-neutral-700 text-neutral-0 rounded-radius-md px-spacing-7 py-spacing-5 text-5 outline-none focus:ring-2 focus:ring-orange-500"
            />
            {errors.password && (
              <p className="text-red-500 text-7">{errors.password.message}</p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="bg-orange-500 hover:bg-orange-400 text-neutral-0 rounded-radius-md text-5-semibold py-spacing-5 mt-spacing-3 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? "Resetting..." : "Reset password"}
          </button>
        </form>

        <p className="mt-spacing-9 text-6 text-neutral-400 text-center">
          <Link to="/login" className="text-orange-500 hover:text-orange-400">
            Back to log in
          </Link>
        </p>
      </div>
    </div>
  );
}
