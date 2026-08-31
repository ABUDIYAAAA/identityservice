import { createTheme } from "@mui/material/styles";

const theme = createTheme({
  palette: {
    mode: "light",
    primary: {
      main: "#0f172a", // Deep sleek slate black
      light: "#1e293b",
      dark: "#020617",
      contrastText: "#ffffff",
    },
    secondary: {
      main: "#475569",
      light: "#64748b",
      dark: "#334155",
      contrastText: "#ffffff",
    },
    background: {
      default: "#f8fafc",
      paper: "#ffffff",
    },
    text: {
      primary: "#0f172a",
      secondary: "#475569",
      disabled: "#94a3b8",
    },
    divider: "#e2e8f0",
  },
  typography: {
    fontFamily: [
      "var(--font-geist-sans)",
      "-apple-system",
      "BlinkMacSystemFont",
      '"Segoe UI"',
      "Roboto",
      '"Helvetica Neue"',
      "Arial",
      "sans-serif",
    ].join(","),
    button: {
      textTransform: "none",
      fontWeight: 600,
    },
  },
  shape: {
    borderRadius: 8,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          boxShadow: "none",
          fontWeight: 600,
          "&:hover": {
            boxShadow: "none",
          },
        },
        containedPrimary: {
          backgroundColor: "#0f172a",
          color: "#ffffff",
          "&:hover": {
            backgroundColor: "#1e293b",
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: "none",
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          boxShadow: "0 1px 3px 0 rgb(0 0 0 / 0.05), 0 1px 2px -1px rgb(0 0 0 / 0.05)",
          borderRadius: 12,
          border: "1px solid #e2e8f0",
        },
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          "&.Mui-focused .MuiOutlinedInput-notchedOutline": {
            borderColor: "#0f172a",
            borderWidth: 1.5,
          },
        },
      },
    },
    MuiInputLabel: {
      styleOverrides: {
        root: {
          "&.Mui-focused": {
            color: "#0f172a",
          },
        },
      },
    },
  },
});

export default theme;
