"use client";

import { useState } from "react";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Box from "@mui/material/Box";
import Alert from "@mui/material/Alert";
import OutlinedInput from "@mui/material/OutlinedInput";
import InputAdornment from "@mui/material/InputAdornment";
import { toast } from "sonner";

export default function SecretDisplayModal({
  open,
  onClose,
  rawSecret,
  clientId,
  serviceName,
}) {
  const [copiedClientId, setCopiedClientId] = useState(false);
  const [copiedSecret, setCopiedSecret] = useState(false);

  const handleCopy = (text, type) => {
    navigator.clipboard.writeText(text);
    if (type === "clientId") {
      setCopiedClientId(true);
      setTimeout(() => setCopiedClientId(false), 2000);
      toast.success("Client ID copied to clipboard");
    } else {
      setCopiedSecret(true);
      setTimeout(() => setCopiedSecret(false), 2000);
      toast.success("Client Secret copied to clipboard");
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      slotProps={{
        paper: {
          elevation: 0,
          sx: {
            borderRadius: 3,
            p: 1,
            border: "1px solid",
            borderColor: "divider",
          },
        },
      }}
    >
      <DialogTitle fontWeight={700}>Generated Secret Credentials</DialogTitle>
      <DialogContent sx={{ pt: 1 }}>
        <Alert
          severity="warning"
          sx={{ mb: 3, borderRadius: 2, fontWeight: 500 }}
        >
          Copy your client secret now. For security, it will{" "}
          <strong>never be shown again</strong>.
        </Alert>

        {serviceName && (
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Credentials for service: <strong>{serviceName}</strong>
          </Typography>
        )}

        {clientId && (
          <Box sx={{ mb: 2.5 }}>
            <Typography
              variant="caption"
              fontWeight={600}
              color="text.secondary"
              display="block"
              sx={{ mb: 0.5 }}
            >
              Client ID
            </Typography>
            <OutlinedInput
              fullWidth
              size="small"
              readOnly
              value={clientId}
              sx={{
                fontFamily: "monospace",
                fontSize: "0.85rem",
                backgroundColor: "#f8fafc",
              }}
              endAdornment={
                <InputAdornment position="end">
                  <Button
                    size="small"
                    onClick={() => handleCopy(clientId, "clientId")}
                    sx={{ textTransform: "none", fontWeight: 600 }}
                  >
                    {copiedClientId ? "Copied!" : "Copy"}
                  </Button>
                </InputAdornment>
              }
            />
          </Box>
        )}

        {rawSecret && (
          <Box sx={{ mb: 1 }}>
            <Typography
              variant="caption"
              fontWeight={600}
              color="text.secondary"
              display="block"
              sx={{ mb: 0.5 }}
            >
              Client Secret
            </Typography>
            <OutlinedInput
              fullWidth
              size="small"
              readOnly
              value={rawSecret}
              sx={{
                fontFamily: "monospace",
                fontSize: "0.875rem",
                fontWeight: 600,
                backgroundColor: "#fefce8",
                borderColor: "#fef08a",
              }}
              endAdornment={
                <InputAdornment position="end">
                  <Button
                    size="small"
                    variant="contained"
                    disableElevation
                    onClick={() => handleCopy(rawSecret, "secret")}
                    sx={{ textTransform: "none", fontWeight: 600 }}
                  >
                    {copiedSecret ? "Copied!" : "Copy Secret"}
                  </Button>
                </InputAdornment>
              }
            />
          </Box>
        )}
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 2.5 }}>
        <Button onClick={onClose} variant="contained" fullWidth sx={{ py: 1 }}>
          I Have Saved This Secret
        </Button>
      </DialogActions>
    </Dialog>
  );
}
