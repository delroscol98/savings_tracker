import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Link, useNavigate } from "react-router";
import { register as apiRegister } from "@/api/auth";
import { handleApiError } from "@/lib/errors";
import { AuthLayout } from "@/components/AuthLayout";

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
    <AuthLayout>
      <div className="pb-10 mt-11">
        <div className="text-neutral-0 grid gap-4">
          <h1 className="text-2">Create your account</h1>
          <p className="text-5 text-neutral-300">
            Start tracking your savings goals
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-8">
        {errors.root && (
          <p className="text-red-500 text-6">{errors.root.message}</p>
        )}

        <div className="flex flex-col gap-5">
          <label htmlFor="full_name" className="text-5 text-neutral-0">
            Full name
          </label>
          <input
            id="full_name"
            type="text"
            {...register("full_name")}
            className="bg-neutral-700 text-neutral-0 rounded-md px-7 py-5 text-5 outline outline-solid outline-neutral-500 focus:ring-2 focus:ring-orange-500"
          />
          {errors.full_name && (
            <p className="text-red-500 text-7">{errors.full_name.message}</p>
          )}
        </div>

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

        <div className="flex flex-col gap-5">
          <label htmlFor="email" className="text-5 text-neutral-0">
            Password
          </label>
          <input
            id="password"
            type="password"
            {...register("password")}
            className="bg-neutral-700 text-neutral-0 rounded-md px-7 py-5 text-5 outline outline-solid outline-neutral-500 focus:ring-2 focus:ring-orange-500"
          />
          {errors.password && (
            <p className="text-red-500 text-7">{errors.password.message}</p>
          )}
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="bg-orange-500 hover:bg-orange-400 text-neutral-0 rounded-full text-5 py-5 mt-6 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? "Creating account..." : "Create account"}
        </button>

        <p className="text-5 text-neutral-400 text-center">
          Already have an account?{" "}
          <Link
            to="/login"
            className="text-neutral-0 hover:text-orange-400 border-b border-neutral-0 hover:border-orange-400"
          >
            Log in
          </Link>
        </p>
      </form>
    </AuthLayout>
  );
}
