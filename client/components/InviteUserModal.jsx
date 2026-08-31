"use client";

import { useState } from "react";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import CircularProgress from "@mui/material/CircularProgress";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { toast } from "sonner";
import api from "@/lib/api";

export default function InviteUserModal({ open, onClose, onSuccess }) {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("user");
  const [submitting, setSubmitting] = useState(false);

  const handleClose = () => {
    if (submitting) return;
    setEmail("");
    setRole("user");
    onClose();
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email) {
      toast.error("Please enter an email address");
      return;
    }

    setSubmitting(true);
    try {
      await api.post("/users/invite", { email, role });
      toast.success(`Invitation sent to ${email}`);
      setEmail("");
      setRole("user");
      onSuccess?.();
      onClose();
    } catch (err) {
      toast.error(err.message || "Failed to send invitation");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog
      open={open}
      onClose={handleClose}
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
      <DialogTitle fontWeight={700}>Invite New User</DialogTitle>
      <Box component="form" onSubmit={handleSubmit}>
        <DialogContent sx={{ pt: 1 }}>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            Send an invitation link for a new team member to activate their
            account.
          </Typography>

          <TextField
            autoFocus
            margin="dense"
            id="invite-email"
            label="Email Address"
            type="email"
            fullWidth
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={submitting}
          />

          <FormControl fullWidth margin="dense" sx={{ mt: 2 }}>
            <InputLabel id="invite-role-label">Assigned Role</InputLabel>
            <Select
              labelId="invite-role-label"
              id="invite-role"
              value={role}
              label="Assigned Role"
              onChange={(e) => setRole(e.target.value)}
              disabled={submitting}
            >
              <MenuItem value="user">User (Standard Access)</MenuItem>
              <MenuItem value="admin">Admin (Full Access)</MenuItem>
            </Select>
          </FormControl>
        </DialogContent>

        <DialogActions sx={{ px: 3, pb: 2.5 }}>
          <Button onClick={handleClose} disabled={submitting} color="inherit">
            Cancel
          </Button>
          <Button
            type="submit"
            variant="contained"
            disabled={submitting}
            sx={{ px: 3 }}
          >
            {submitting ? (
              <CircularProgress size={20} color="inherit" />
            ) : (
              "Send Invitation"
            )}
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  );
}
