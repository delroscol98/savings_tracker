import type { ReactNode } from "react";

import logo from "../assets/logo-large.svg";
import star from "../assets/pattern-star.svg";

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-neutral-900 px-7 md:p-13 lg:px-14 lg:py-11 grid lg:grid-cols-2 items-center content-center">
      <div
        className={`sm:hidden md:hidden lg:min-h-full lg:grid lg:grid-rows-[1fr_min-content] lg:mr-14 lg:px-11 lg:py-8 lg:rounded-2xl lg:bg-linear-to-br lg:from-orange-700 lg:to-orange-400 lg:relative lg:overflow-hidden`}
      >
        <img
          className="sm:hidden md:hidden lg:inline lg:absolute lg:-right-15 lg:-bottom-17 lg:w-3/4"
          src={star}
          alt="logo"
        />
        <span className="text-1 text-neutral-0 self-center">
          "The goal isn't to be rich. It's to have enough."
        </span>
        <span className="text-4 text-neutral-0">- Morgan Housel</span>
      </div>
      <div>
        <img src={logo} alt="logo" />

        {children}
      </div>
    </div>
  );
}
