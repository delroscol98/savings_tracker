import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Link } from "react-router";
import { forgotPassword } from "@/api/auth";
import { handleApiError } from "@/lib/errors";

const forgotPasswordSchema = z.object({
  email: z.string().email("Please enter a valid email"),
});

type ForgotPasswordForm = z.infer<typeof forgotPasswordSchema>;

export function ForgotPasswordPage() {
  const [sent, setSent] = useState(false);
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordForm>({
    resolver: zodResolver(forgotPasswordSchema),
  });

  async function onSubmit(data: ForgotPasswordForm) {
    try {
      await forgotPassword(data.email);
      setSent(true);
    } catch (error: unknown) {
      handleApiError(error, setError);
    }
  }

  if (sent) {
    return (
      <div className="min-h-screen bg-neutral-900 flex items-center justify-center px-spacing-7">
        <div className="bg-neutral-800 rounded-radius-2xl p-spacing-12 w-full max-w-[400px]">
          <h1 className="text-3 text-neutral-0 mb-spacing-7">Check your email</h1>
          <p className="text-5 text-neutral-300 mb-spacing-9">
            If an account exists with that email, we&apos;ve sent a password reset
            link.
          </p>
          <Link
            to="/login"
            className="block text-center text-6 text-orange-500 hover:text-orange-400"
          >
            Back to log in
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-neutral-900 flex items-center justify-center px-spacing-7">
      <div className="bg-neutral-800 rounded-radius-2xl p-spacing-12 w-full max-w-[400px]">
        <h1 className="text-3 text-neutral-0 mb-spacing-10">Forgot password</h1>

        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-spacing-7">
          {errors.root && (
            <p className="text-red-500 text-6">{errors.root.message}</p>
          )}

          <div className="flex flex-col gap-spacing-3">
            <label htmlFor="email" className="text-6 text-neutral-300">
              Email
            </label>
            <input
              id="email"
              type="email"
              {...register("email")}
              className="bg-neutral-700 text-neutral-0 rounded-radius-md px-spacing-7 py-spacing-5 text-5 outline-none focus:ring-2 focus:ring-orange-500"
            />
            {errors.email && (
              <p className="text-red-500 text-7">{errors.email.message}</p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="bg-orange-500 hover:bg-orange-400 text-neutral-0 rounded-radius-md text-5-semibold py-spacing-5 mt-spacing-3 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? "Sending..." : "Send reset link"}
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
