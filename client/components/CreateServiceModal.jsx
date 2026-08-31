"use client";

import { useState } from "react";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { toast } from "sonner";
import api from "@/lib/api";

export default function CreateServiceModal({ open, onClose, onSuccess }) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleClose = () => {
    if (submitting) return;
    setName("");
    setDescription("");
    onClose();
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name) {
      toast.error("Service name is required");
      return;
    }

    setSubmitting(true);
    try {
      const res = await api.post("/services", { name, description });
      toast.success("Service created successfully!");
      const data = res.data?.data;
      setName("");
      setDescription("");
      onClose();
      onSuccess?.(data);
    } catch (err) {
      toast.error(err.message || "Failed to create service");
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
      <DialogTitle fontWeight={700}>Create New Service</DialogTitle>
      <Box component="form" onSubmit={handleSubmit}>
        <DialogContent sx={{ pt: 1 }}>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            Register a new service application to generate Client ID and API
            access credentials.
          </Typography>

          <TextField
            autoFocus
            margin="dense"
            id="service-name"
            label="Service Name"
            fullWidth
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={submitting}
            placeholder="e.g. Auth Gateway, Billing Service"
          />

          <TextField
            margin="dense"
            id="service-desc"
            label="Description (Optional)"
            fullWidth
            multiline
            rows={3}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={submitting}
            sx={{ mt: 2 }}
          />
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
              "Create Service"
            )}
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  );
}
