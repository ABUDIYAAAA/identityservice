"use client";

import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";

export default function ConfirmDeleteDialog({
  open,
  onClose,
  onConfirm,
  targetName,
  title,
  description,
  confirmText,
  loading,
}) {
  const dialogTitle = title || "Confirm Delete";
  const buttonText = confirmText || title || "Delete";

  return (
    <Dialog
      open={open}
      onClose={loading ? undefined : onClose}
      maxWidth="xs"
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
      <DialogTitle fontWeight={700} color="error.main">
        {dialogTitle}
      </DialogTitle>
      <DialogContent sx={{ pt: 1 }}>
        <Typography variant="body2" color="text.secondary">
          {description || (
            <>
              Are you sure you want to delete <strong>{targetName}</strong>? This action cannot be undone.
            </>
          )}
        </Typography>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2.5 }}>
        <Button onClick={onClose} disabled={loading} color="inherit">
          Cancel
        </Button>
        <Button
          onClick={onConfirm}
          variant="contained"
          color="error"
          disabled={loading}
          sx={{ px: 3 }}
        >
          {loading ? <CircularProgress size={20} color="inherit" /> : buttonText}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
