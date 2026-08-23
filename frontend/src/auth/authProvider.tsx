import { login as apiLogin } from "@/api/auth";
import { useCallback, useEffect, useReducer, type ReactNode } from "react";
import {
  AuthCtx,
  initialAuthState,
  type AuthAction,
  type AuthState,
} from "./authContext";
import {
  clearStoredToken,
  clearStoredUser,
  getStoredToken,
  getStoredUser,
  setStoredToken,
  setStoredUser,
} from "@/api/localStorage";
import { ApiError, setOnUnauthorized } from "@/api/client";
import type { User } from "@/api/schemas";

function authReducer(state: AuthState, action: AuthAction) {
  switch (action.type) {
    case "LOGIN_START":
      return { ...state, isLoading: true };
    case "LOGIN_SUCCESS":
      return {
        user: action.payload.user,
        token: action.payload.token,
        isLoading: false,
      };
    case "LOGIN_FAILURE":
      return { user: null, token: null, isLoading: false };
    case "LOGOUT":
      return { user: null, token: null, isLoading: false };
    case "HYDRATE":
      return {
        user: action.payload.user,
        token: action.payload.token,
        isLoading: false,
      };
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [{ user, token, isLoading }, dispatch] = useReducer(
    authReducer,
    initialAuthState,
  );

  useEffect(() => {
    const token = getStoredToken();
    if (token == "") {
      return;
    }

    const storedUser = getStoredUser();
    if (storedUser) {
      try {
        const user = storedUser;
        dispatch({ type: "HYDRATE", payload: { user, token } });
      } catch {
        clearStoredToken();
      }
    }
  }, []);

  const logout = useCallback(() => {
    clearStoredToken();
    clearStoredUser();
    dispatch({ type: "LOGOUT" });
  }, []);

  useEffect(() => {
    setOnUnauthorized(logout);
    return () => setOnUnauthorized(null);
  }, [logout]);

  async function login(email: string, password: string) {
    dispatch({ type: "LOGIN_START" });
    try {
      const response = await apiLogin(email, password);
      const user: User = {
        id: response.id,
        created_at: response.created_at,
        updated_at: response.updated_at,
        email: response.email,
      };

      setStoredToken(response.token);
      dispatch({
        type: "LOGIN_SUCCESS",
        payload: {
          user: user,
          token: response.token,
        },
      });
      setStoredUser(user as User);
    } catch (error: unknown) {
      dispatch({ type: "LOGIN_FAILURE" });
      if (error instanceof ApiError) {
        throw new Error(error.message, { cause: error });
      }
    }
  }

  const value = {
    user,
    token,
    isLoading,
    login,
    logout,
  };

  return <AuthCtx.Provider value={value}>{children}</AuthCtx.Provider>;
}
