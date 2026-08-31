"use client";

import { useState, useEffect, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import Link from "next/link";
import Container from "@mui/material/Container";
import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Chip from "@mui/material/Chip";
import Alert from "@mui/material/Alert";
import { toast } from "sonner";
import { useAuth } from "@/hooks/useAuth";

function AcceptInviteForm() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const tokenFromUrl = searchParams.get("token") || "";

  const [token, setToken] = useState(tokenFromUrl);
  const [inviteDetails, setInviteDetails] = useState(null);
  const [fetchingDetails, setFetchingDetails] = useState(Boolean(tokenFromUrl));
  const [inviteError, setInviteError] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const { getInviteDetails, acceptInvite } = useAuth();

  useEffect(() => {
    if (!tokenFromUrl) return;

    let isMounted = true;
    setFetchingDetails(true);

    getInviteDetails(tokenFromUrl)
      .then((details) => {
        if (isMounted) {
          setInviteDetails(details);
          setInviteError("");
        }
      })
      .catch((err) => {
        if (isMounted) {
          setInviteError(
            err.message || "Invitation link is invalid or expired.",
          );
        }
      })
      .finally(() => {
        if (isMounted) setFetchingDetails(false);
      });

    return () => {
      isMounted = false;
    };
  }, [tokenFromUrl, getInviteDetails]);

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!token) {
      toast.error("Invitation token is required");
      return;
    }
    if (!password || password.length < 8) {
      toast.error("Password must be at least 8 characters long");
      return;
    }
    if (password !== confirmPassword) {
      toast.error("Passwords do not match");
      return;
    }

    setSubmitting(true);
    try {
      await acceptInvite(token, password, confirmPassword);
      toast.success("Account activated successfully! Please sign in.");
      router.push("/login");
    } catch (err) {
      toast.error(err.message || "Failed to accept invitation");
    } finally {
      setSubmitting(false);
    }
  };

  if (fetchingDetails) {
    return (
      <Box sx={{ textAlign: "center", py: 4 }}>
        <CircularProgress size={32} />
        <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
          Validating invitation...
        </Typography>
      </Box>
    );
  }

  if (inviteError) {
    return (
      <Box sx={{ textAlign: "center" }}>
        <Alert severity="error" sx={{ mb: 3, borderRadius: 2 }}>
          {inviteError}
        </Alert>
        <Button component={Link} href="/login" fullWidth variant="outlined">
          Go to Sign In
        </Button>
      </Box>
    );
  }

  return (
    <Box component="form" onSubmit={handleSubmit} noValidate>
      {inviteDetails && (
        <Box
          sx={{
            p: 2,
            mb: 3,
            backgroundColor: "background.default",
            borderRadius: 2,
            border: "1px solid",
            borderColor: "divider",
          }}
        >
          <Typography variant="caption" color="text.secondary" display="block">
            Invited Email
          </Typography>
          <Typography variant="body1" fontWeight={600} gutterBottom>
            {inviteDetails.email}
          </Typography>

          <Box sx={{ display: "flex", alignItems: "center", gap: 1, mt: 1 }}>
            <Typography variant="caption" color="text.secondary">
              Assigned Role:
            </Typography>
            <Chip
              label={inviteDetails.role}
              size="small"
              sx={{ fontWeight: 600, textTransform: "capitalize" }}
            />
          </Box>
        </Box>
      )}

      {!tokenFromUrl && (
        <TextField
          margin="normal"
          required
          fullWidth
          id="token"
          label="Invitation Token"
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
        name="password"
        label="Set Password"
        type="password"
        id="password"
        autoFocus
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        disabled={submitting}
        helperText="Minimum 8 characters"
      />

      <TextField
        margin="normal"
        required
        fullWidth
        name="confirmPassword"
        label="Confirm Password"
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
          "Activate Account"
        )}
      </Button>

      <Box sx={{ textAlign: "center" }}>
        <Link href="/login" style={{ textDecoration: "none" }}>
          <Typography
            variant="body2"
            color="text.secondary"
            sx={{ "&:hover": { textDecoration: "underline" } }}
          >
            Already registered? Sign in
          </Typography>
        </Link>
      </Box>
    </Box>
  );
}

export default function AcceptInvitePage() {
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
              Accept Invitation
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Set up your password to activate your account
            </Typography>
          </Box>

          <Suspense
            fallback={
              <Box sx={{ textAlign: "center", py: 4 }}>
                <CircularProgress />
              </Box>
            }
          >
            <AcceptInviteForm />
          </Suspense>
        </Paper>
      </Container>
    </Box>
  );
}
