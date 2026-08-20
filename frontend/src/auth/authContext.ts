import { login } from "@/api/auth";
import { clearStoredToken, client } from "@/api/client";
import type { LoginResponse, User } from "@/api/schemas";
import { createContext, useReducer } from "react";

type AuthContextValue = {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<LoginResponse>;
  logout: () => void;
};

const AuthContext = createContext(null);

const initialState: AuthContextValue = {
  user: null,
  token: null,
  isLoading: false,
  login: login,
  logout: function logout() {
    this.user = null;
    this.token = null;

    clearStoredToken();
  },
};
