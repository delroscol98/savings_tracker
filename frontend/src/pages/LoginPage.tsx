import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Link, useNavigate } from "react-router";
import { useAuth } from "@/auth/authContext";
import { handleApiError } from "@/lib/errors";

const loginSchema = z.object({
  email: z.string().email("Please enter a valid email"),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters"),
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
    <div className="min-h-screen bg-neutral-900 flex items-center justify-center px-spacing-7">
      <div className="bg-neutral-800 rounded-radius-2xl p-spacing-12 w-full max-w-[400px]">
        <h1 className="text-3 text-neutral-0 mb-spacing-10">Log in</h1>

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

          <div className="flex flex-col gap-spacing-3">
            <label htmlFor="password" className="text-6 text-neutral-300">
              Password
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
            {isSubmitting ? "Logging in..." : "Log in"}
          </button>
        </form>

        <div className="mt-spacing-9 flex flex-col items-center gap-spacing-4">
          <Link
            to="/forgot-password"
            className="text-6 text-neutral-400 hover:text-orange-500"
          >
            Forgot password?
          </Link>
          <p className="text-6 text-neutral-400">
            Don&apos;t have an account?{" "}
            <Link to="/signup" className="text-orange-500 hover:text-orange-400">
              Sign up
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
