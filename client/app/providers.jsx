"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { AppRouterCacheProvider } from "@mui/material-nextjs/v15-appRouter";
import { ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import { Toaster } from "sonner";
import theme from "@/lib/theme";
import { AuthProvider } from "@/context/AuthContext";

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 0, // Data is considered stale immediately for background updates
        refetchOnWindowFocus: true, // Refetch automatically when window gains focus
        refetchOnReconnect: true, // Refetch when network reconnects
        retry: 1,
      },
    },
  });
}

let browserQueryClient;

function getQueryClient() {
  if (typeof window === "undefined") {
    return makeQueryClient();
  } else {
    if (!browserQueryClient) browserQueryClient = makeQueryClient();
    return browserQueryClient;
  }
}

export default function Providers({ children }) {
  const queryClient = getQueryClient();

  return (
    <AppRouterCacheProvider options={{ enableCssLayer: true }}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <QueryClientProvider client={queryClient}>
          <AuthProvider>
            {children}
            <Toaster position="top-right" richColors closeButton />
            <ReactQueryDevtools initialIsOpen={false} />
          </AuthProvider>
        </QueryClientProvider>
      </ThemeProvider>
    </AppRouterCacheProvider>
  );
}
