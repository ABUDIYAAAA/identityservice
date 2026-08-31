"use client";

import { useState } from "react";
import Link from "next/link";
import Container from "@mui/material/Container";
import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Alert from "@mui/material/Alert";
import { toast } from "sonner";
import { useAuth } from "@/hooks/useAuth";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const { forgotPassword } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email) {
      toast.error("Please enter your email address");
      return;
    }

    setSubmitting(true);
    try {
      await forgotPassword(email);
      setSubmitted(true);
      toast.success("Password reset request sent");
    } catch (err) {
      toast.error(err.message || "Failed to send reset link");
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
              Forgot Password
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Enter your email address to receive password reset instructions
            </Typography>
          </Box>

          {submitted ? (
            <Box sx={{ textAlign: "center" }}>
              <Alert severity="success" sx={{ mb: 3, borderRadius: 2 }}>
                If an account exists for <strong>{email}</strong>, a password
                reset link has been sent.
              </Alert>
              <Button
                component={Link}
                href="/login"
                fullWidth
                variant="outlined"
                sx={{ py: 1 }}
              >
                Back to Sign In
              </Button>
            </Box>
          ) : (
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

              <Button
                type="submit"
                fullWidth
                variant="contained"
                size="large"
                disabled={submitting}
                sx={{ py: 1.2, mt: 3, mb: 2 }}
              >
                {submitting ? (
                  <CircularProgress size={24} color="inherit" />
                ) : (
                  "Send Reset Link"
                )}
              </Button>

              <Box sx={{ textAlign: "center" }}>
                <Link href="/login" style={{ textDecoration: "none" }}>
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ "&:hover": { textDecoration: "underline" } }}
                  >
                    Remember your password? Sign in
                  </Typography>
                </Link>
              </Box>
            </Box>
          )}
        </Paper>
      </Container>
    </Box>
  );
}
