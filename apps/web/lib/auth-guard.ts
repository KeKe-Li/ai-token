"use client";

import { useEffect, useState } from "react";

export function useAuthGuard(requireAuth: boolean) {
  const [checked, setChecked] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("token");
    const hasAuth = !!token;
    setIsLoggedIn(hasAuth);

    if (requireAuth && !hasAuth) {
      window.location.href = "/login";
      return;
    }

    if (!requireAuth && hasAuth) {
      window.location.href = "/dashboard";
      return;
    }

    setChecked(true);
  }, [requireAuth]);

  return { checked, isLoggedIn };
}
