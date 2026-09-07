import { useSearchParams, Link, useNavigate } from "react-router";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { resetPassword } from "@/api/auth";
import { handleApiError } from "@/lib/errors";
import { AuthLayout } from "@/components/AuthLayout";
import { useState } from "react";

const resetPasswordSchema = z.object({
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(128, "Password must be at most 128 characters"),
});

type ResetPasswordForm = z.infer<typeof resetPasswordSchema>;

export function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const [newPassword, setNewPassword] = useState("");
  const [confirmNewPassword, setConfirmNewPassword] = useState("");
  const [clientError, setClientError] = useState("");
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
      <AuthLayout>
        <div className="pb-11 mt-11">
          <div className="text-neutral-0 grid gap-7">
            <h1 className="text-2">Invalid reset link</h1>
            <p className="text-5 text-neutral-300">
              This password reset link is invalid or expired
            </p>
          </div>
        </div>

        <Link
          to="/forgot-password"
          className="inline text-6 text-neutral-0 hover:text-orange-400 border-b border-neutral-0 hover:border-orange-400"
        >
          Request a new reset link
        </Link>
      </AuthLayout>
    );
  }

  async function onSubmit(data: ResetPasswordForm) {
    if (newPassword != confirmNewPassword) {
      setClientError("Passwords do not match");
      return;
    }

    try {
      await resetPassword(token!, data.password);
      navigate("/login");
    } catch (error: unknown) {
      handleApiError(error, setError);
    }
  }

  return (
    <AuthLayout>
      <div className="pb-11 mt-11">
        <div className="text-neutral-0 grid gap-7">
          <h1 className="text-2">Create new password</h1>
          <p className="text-5 text-neutral-300">
            Your new password must be different from your previous password
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-8">
        {errors.root && (
          <p className="text-red-500 text-6">{errors.root.message}</p>
        )}

        <div className="flex flex-col gap-5">
          <label htmlFor="full_name" className="text-5 text-neutral-0">
            New Password
          </label>
          <input
            id="full_name"
            type="text"
            {...register("password")}
            className="bg-neutral-700 text-neutral-0 rounded-md px-7 py-5 text-5 outline outline-solid outline-neutral-500 focus:ring-2 focus:ring-orange-500"
            value={newPassword}
            onChange={(e) => {
              setNewPassword(e.target.value);
            }}
          />
          <span className="text-6 text-neutral-300">At least 8 characters</span>
          {errors.password && (
            <p className="text-red-500 text-7">{errors.password.message}</p>
          )}
        </div>

        <div className="flex flex-col gap-5">
          <label htmlFor="full_name" className="text-5 text-neutral-0">
            Confirm new Password
          </label>
          <input
            id="full_name"
            type="text"
            value={confirmNewPassword}
            onChange={(e) => {
              setConfirmNewPassword(e.target.value);
            }}
            onBlur={() => {
              if (newPassword != confirmNewPassword) {
                setClientError("Passwords do not match");
              } else {
                setClientError("");
              }
            }}
            className="bg-neutral-700 text-neutral-0 rounded-md px-7 py-5 text-5 outline outline-solid outline-neutral-500 focus:ring-2 focus:ring-orange-500"
          />
          {clientError && <p className="text-red-500 text-7">{clientError}</p>}
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="bg-orange-500 hover:bg-orange-400 text-neutral-900 rounded-full text-5 py-5 mt-3 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? "Resetting..." : "Reset password"}
        </button>

        <p className="mt-spacing-9 text-6 text-neutral-400 text-center">
          <Link
            to="/login"
            className="text-neutral-0 hover:text-orange-400 border-b border-neutral-0 hover:border-orange-400"
          >
            Back to sign in
          </Link>
        </p>
      </form>
    </AuthLayout>
  );
}
