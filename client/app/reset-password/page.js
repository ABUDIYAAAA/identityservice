"use client";

import { useState, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
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

function ResetPasswordForm() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const tokenFromUrl = searchParams.get("token") || "";

  const [token, setToken] = useState(tokenFromUrl);
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const { resetPassword } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!token) {
      toast.error("Reset token is required");
      return;
    }
    if (!newPassword || newPassword.length < 8) {
      toast.error("Password must be at least 8 characters long");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("Passwords do not match");
      return;
    }

    setSubmitting(true);
    try {
      await resetPassword(token, newPassword, confirmPassword);
      toast.success("Password has been reset successfully. Please log in.");
      router.push("/login");
    } catch (err) {
      toast.error(err.message || "Failed to reset password");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Box component="form" onSubmit={handleSubmit} noValidate>
      {!tokenFromUrl && (
        <TextField
          margin="normal"
          required
          fullWidth
          id="token"
          label="Reset Token"
          name="token"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          disabled={submitting}
        />
      )}

      <TextField
        margin="normal"
        required
        fullWidth
        name="newPassword"
        label="New Password"
        type="password"
        id="newPassword"
        autoFocus={Boolean(tokenFromUrl)}
        value={newPassword}
        onChange={(e) => setNewPassword(e.target.value)}
        disabled={submitting}
        helperText="Minimum 8 characters"
      />

      <TextField
        margin="normal"
        required
        fullWidth
        name="confirmPassword"
        label="Confirm New Password"
        type="password"
        id="confirmPassword"
        value={confirmPassword}
        onChange={(e) => setConfirmPassword(e.target.value)}
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
          "Reset Password"
        )}
      </Button>

      <Box sx={{ textAlign: "center" }}>
        <Link href="/login" style={{ textDecoration: "none" }}>
          <Typography
            variant="body2"
            color="text.secondary"
            sx={{ "&:hover": { textDecoration: "underline" } }}
          >
            Back to Sign In
          </Typography>
        </Link>
      </Box>
    </Box>
  );
}

export default function ResetPasswordPage() {
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
              Set New Password
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Enter your new password below
            </Typography>
          </Box>

          <Suspense
            fallback={
              <Box sx={{ textAlign: "center", py: 4 }}>
                <CircularProgress />
              </Box>
            }
          >
            <ResetPasswordForm />
          </Suspense>
        </Paper>
      </Container>
    </Box>
  );
}
