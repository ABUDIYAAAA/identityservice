"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Container from "@mui/material/Container";
import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import { toast } from "sonner";
import { useAuth } from "@/hooks/useAuth";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email || !password) {
      toast.error("Please fill in all fields");
      return;
    }

    setSubmitting(true);
    try {
      await login(email, password);
      toast.success("Successfully logged in");
      router.push("/");
    } catch (err) {
      toast.error(err.message || "Failed to log in");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        backgroundColor: "background.default",
        py: 4,
      }}
    >
      <Container maxWidth="xs">
        <Paper
          elevation={0}
          sx={{
            p: 4,
            borderRadius: 3,
            border: "1px solid",
            borderColor: "divider",
          }}
        >
          <Box sx={{ mb: 3, textAlign: "center" }}>
            <Typography
              variant="h5"
              component="h1"
              fontWeight={700}
              gutterBottom
            >
              Sign In
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Enter your credentials to access your account
            </Typography>
          </Box>

          <Box component="form" onSubmit={handleSubmit} noValidate>
            <TextField
              margin="normal"
              required
              fullWidth
              id="email"
              label="Email Address"
              name="email"
              autoComplete="email"
              autoFocus
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={submitting}
            />

            <TextField
              margin="normal"
              required
              fullWidth
              name="password"
              label="Password"
              type="password"
              id="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={submitting}
            />

            <Box
              sx={{ display: "flex", justifyContent: "flex-end", mt: 1, mb: 2 }}
            >
              <Link href="/forgot-password" style={{ textDecoration: "none" }}>
                <Typography
                  variant="body2"
                  color="primary"
                  sx={{
                    fontWeight: 500,
                    "&:hover": { textDecoration: "underline" },
                  }}
                >
                  Forgot password?
                </Typography>
              </Link>
            </Box>

            <Button
              type="submit"
              fullWidth
              variant="contained"
              size="large"
              disabled={submitting}
              sx={{ py: 1.2, mt: 1 }}
            >
              {submitting ? (
                <CircularProgress size={24} color="inherit" />
              ) : (
                "Sign In"
              )}
            </Button>
          </Box>
        </Paper>
      </Container>
    </Box>
  );
}
