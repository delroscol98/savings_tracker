import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Link, useNavigate } from "react-router";
import { useAuth } from "@/auth/authContext";
import { handleApiError } from "@/lib/errors";
import { AuthLayout } from "@/components/AuthLayout";

const loginSchema = z.object({
  email: z.email("Please enter a valid email"),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

type LoginForm = z.infer<typeof loginSchema>;

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  });

  async function onSubmit(data: LoginForm) {
    try {
      await login(data.email, data.password);
      navigate("/");
    } catch (error: unknown) {
      handleApiError(error, setError);
    }
  }

  return (
    <AuthLayout>
      <div className="pb-10 mt-11 border-b border-neutral-700">
        <div className="text-neutral-0 grid gap-4">
          <h1 className="text-2">Welcome back</h1>
          <p className="text-5 text-neutral-300">Sign into your account</p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="grid gap-8 mt-10">
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
            className="bg-neutral-700 text-neutral-0 rounded-md px-7 py-5 text-5 outline outline-solid outline-neutral-500 focus:ring-2 focus:ring-orange-500"
          />
          {errors.email && (
            <p className="text-red-500 text-7">{errors.email.message}</p>
          )}
        </div>

        <div className="grid gap-6">
          <div className="flex flex-col gap-5">
            <label htmlFor="password" className="text-5 text-neutral-0">
              Password
            </label>
            <input
              id="password"
              type="password"
              {...register("password")}
              className="bg-neutral-700 text-neutral-0 rounded-md px-7 py-5 text-5 outline outline-solid outline-neutral-500 outline-none focus:ring-2 focus:ring-orange-500"
            />
            {errors.password && (
              <p className="text-red-500 text-7">{errors.password.message}</p>
            )}
          </div>
          <Link
            to="/forgot-password"
            className="text-5 text-neutral-300 hover:text-orange-500 justify-self-end"
          >
            Forgot password?
          </Link>
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="bg-orange-500 hover:bg-orange-400 text-neutral-900 rounded-full text-5-semibold py-5 mt-3 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? "Logging in..." : "Sign in"}
        </button>

        <div className="mt-spacing-9 flex flex-col items-center gap-spacing-4">
          <p className="text-5 text-neutral-400">
            Don&apos;t have an account?{" "}
            <Link
              to="/signup"
              className="text-5 text-neutral-0 border-b border-neutal-0 hover:text-orange-400"
            >
              Create one
            </Link>
          </p>
        </div>
      </form>
    </AuthLayout>
  );
}
