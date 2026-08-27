import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Link, useNavigate } from "react-router";
import { register as apiRegister } from "@/api/auth";
import { handleApiError } from "@/lib/errors";

const registerSchema = z.object({
  full_name: z.string().min(1, "Name is required"),
  email: z.string().email("Please enter a valid email"),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(128, "Password must be at most 128 characters"),
});

type RegisterForm = z.infer<typeof registerSchema>;

export function SignUpPage() {
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  });

  async function onSubmit(data: RegisterForm) {
    try {
      await apiRegister(data.email, data.password, data.full_name);
      navigate("/login");
    } catch (error: unknown) {
      handleApiError(error, setError);
    }
  }

  return (
    <div className="min-h-screen bg-neutral-900 flex items-center justify-center px-spacing-7">
      <div className="bg-neutral-800 rounded-radius-2xl p-spacing-12 w-full max-w-[400px]">
        <h1 className="text-3 text-neutral-0 mb-spacing-10">Sign up</h1>

        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-spacing-7">
          {errors.root && (
            <p className="text-red-500 text-6">{errors.root.message}</p>
          )}

          <div className="flex flex-col gap-spacing-3">
            <label htmlFor="full_name" className="text-6 text-neutral-300">
              Full name
            </label>
            <input
              id="full_name"
              type="text"
              {...register("full_name")}
              className="bg-neutral-700 text-neutral-0 rounded-radius-md px-spacing-7 py-spacing-5 text-5 outline-none focus:ring-2 focus:ring-orange-500"
            />
            {errors.full_name && (
              <p className="text-red-500 text-7">{errors.full_name.message}</p>
            )}
          </div>

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
            {isSubmitting ? "Creating account..." : "Create account"}
          </button>
        </form>

        <p className="mt-spacing-9 text-6 text-neutral-400 text-center">
          Already have an account?{" "}
          <Link to="/login" className="text-orange-500 hover:text-orange-400">
            Log in
          </Link>
        </p>
      </div>
    </div>
  );
}
