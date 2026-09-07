import type { ReactNode } from "react";

import logo from "../assets/logo-large.svg";
import star from "../assets/pattern-star.svg";
import { useLocation } from "react-router";
import { Link } from "react-router";

export function AuthLayout({ children }: { children: ReactNode }) {
  const location = useLocation();
  function quoteSetter() {
    switch (location.pathname) {
      case "/login":
        return (
          <>
            <span className="text-1 text-neutral-0 self-center">
              "The goal isn't to be rich. It's to have enough."
            </span>
            <span className="text-4 text-neutral-0">- Morgan Housel</span>
          </>
        );
      case "/signup":
        return (
          <>
            <span className="text-1 text-neutral-0 self-center">
              "Do not save what is left after spending, but spend what is left
              after saving"
            </span>
            <span className="text-4 text-neutral-0">- Warren Buffet</span>
          </>
        );
      case "/forgot-password":
      case "/reset-password":
        return (
          <>
            <span className="text-1 text-neutral-0 self-center">
              "A budget is telling your money where to go instead of wondering
              where it went"
            </span>
            <span className="text-4 text-neutral-0">- Dave Ramsey</span>
          </>
        );
    }
  }
  return (
    <div className="min-h-screen bg-neutral-900 px-7 md:p-13 lg:px-14 lg:py-11 grid lg:grid-cols-2 items-center">
      <div
        className={`max-lg:hidden lg:min-h-full lg:grid lg:grid-rows-[1fr_min-content] lg:mr-14 lg:px-11 lg:py-8 lg:rounded-2xl lg:bg-linear-to-br lg:from-orange-700 lg:to-orange-400 lg:relative lg:overflow-hidden`}
      >
        <img
          className="max-lg:hidden lg:inline lg:absolute lg:-right-15 lg:-bottom-17 lg:w-3/4"
          src={star}
          alt="logo"
        />
        {quoteSetter()}
      </div>
      <div>
        <Link to="/login">
          <img src={logo} alt="logo" />
        </Link>

        {children}
      </div>
    </div>
  );
}
