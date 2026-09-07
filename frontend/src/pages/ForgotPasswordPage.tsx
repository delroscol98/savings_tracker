import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Link } from "react-router";
import { forgotPassword } from "@/api/auth";
import { handleApiError } from "@/lib/errors";
import { AuthLayout } from "@/components/AuthLayout";
import chevron from "../assets/icon-chevron-left.svg";
import { toast } from "sonner";

const forgotPasswordSchema = z.object({
  email: z.email("Please enter a valid email"),
});

type ForgotPasswordForm = z.infer<typeof forgotPasswordSchema>;

export function ForgotPasswordPage() {
  const [sent, setSent] = useState(false);
  const [email, setEmail] = useState("");
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordForm>({
    resolver: zodResolver(forgotPasswordSchema),
  });

  async function onSubmit(data: ForgotPasswordForm) {
    const toastId = toast.loading("Sending...");
    try {
      await forgotPassword(data.email);
      toast.success("Email sent!", { id: toastId });
      setSent(true);
    } catch (error: unknown) {
      handleApiError(error, setError, toastId);
    }
  }

  if (sent) {
    return (
      <AuthLayout>
        <div className="grid gap-10 mt-11">
          <div className="grid gap-4">
            <h1 className="text-2 text-neutral-0">Check your inbox</h1>
            <p className="text-5 text-neutral-300">
              We've sent a reset link to{" "}
              <span className="text-neutral-0">{email}</span>
            </p>
          </div>
          <p className="text-5 text-neutral-0">
            This link expires in 30 minutes
          </p>
          <div className="grid gap-7">
            <p className="text-5 text-neutral-300">
              Didn't receive it?{" "}
              <span
                className="text-neutral-0 hover:text-orange-400 border-b border-neutral-0 hover:border-orange-400 cursor-pointer"
                onClick={handleSubmit(onSubmit)}
              >
                Resend email
              </span>
            </p>
            <Link to="/login" className="text-5 text-neutral-300 flex gap-3">
              <img src={chevron} alt="" />
              Back to sign in
            </Link>
          </div>
        </div>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout>
      <div className="pb-10 mt-11">
        <div className="text-neutral-0 grid gap-7">
          <h1 className="text-2">Forgot your password</h1>
          <p className="text-5 text-neutral-300">
            Enter your email address and we'll send you a link to reset it
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-8">
        {errors.root && (
          <p className="text-red-500 text-6">{errors.root.message}</p>
        )}

        <div className="flex flex-col gap-5">
          <label htmlFor="email" className="text-5 text-neutral-0">
            Email address
          </label>
          <input
            id="email"
            type="email"
            {...register("email")}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="bg-neutral-700 text-neutral-0 rounded-md px-7 py-5 text-5 outline outline-solid outline-neutral-500 focus:ring-2 focus:ring-orange-500"
          />
          {errors.email && (
            <p className="text-red-500 text-7">{errors.email.message}</p>
          )}
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="bg-orange-500 hover:bg-orange-400 text-neutral-900 rounded-full text-5 py-5 mt-3 disabled:opacity-50 cursor-pointer"
        >
          {isSubmitting ? "Sending..." : "Send reset link"}
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
