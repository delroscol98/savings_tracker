import type { User } from "@/api/schemas";
import { createContext, useContext } from "react";

export type AuthState = {
  user: User | null;
  token: string | null;
  isLoading: boolean;
};

export type AuthContextValue = AuthState & {
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
};

export type AuthPayload = {
  user: User;
  token: string;
};

export type AuthAction =
  | { type: "LOGIN_START" }
  | { type: "LOGIN_SUCCESS"; payload: AuthPayload }
  | { type: "LOGIN_FAILURE" }
  | { type: "LOGOUT" }
  | { type: "HYDRATE"; payload: AuthPayload };

export const initialAuthState: AuthState = {
  user: null,
  token: null,
  isLoading: false,
};

export const AuthCtx = createContext<AuthContextValue | null>(null);

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthCtx);

  if (ctx == null) {
    throw new Error("Auth Context must be used inside Auth Provider");
  }

  return ctx;
}
