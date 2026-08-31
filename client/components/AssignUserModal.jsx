"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Button from "@mui/material/Button";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import CircularProgress from "@mui/material/CircularProgress";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Alert from "@mui/material/Alert";
import { toast } from "sonner";
import api from "@/lib/api";

export default function AssignUserModal({
  open,
  onClose,
  serviceId,
  serviceName,
  onSuccess,
}) {
  const [selectedUserId, setSelectedUserId] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const {
    data: usersData,
    isLoading: loadingUsers,
    isError,
    error,
  } = useQuery({
    queryKey: ["all-users-for-assignment"],
    queryFn: async () => {
      const res = await api.get("/users", { params: { limit: 100 } });
      return res.data?.data?.users || [];
    },
    enabled: Boolean(open),
  });

  const handleClose = () => {
    if (submitting) return;
    setSelectedUserId("");
    onClose();
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!selectedUserId) {
      toast.error("Please select a user to assign");
      return;
    }

    setSubmitting(true);
    try {
      await api.post(`/services/${serviceId}/assign`, {
        user_id: selectedUserId,
      });
      toast.success("User assigned to service successfully");
      setSelectedUserId("");
      onSuccess?.();
      onClose();
    } catch (err) {
      toast.error(err.message || "Failed to assign user");
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
      <DialogTitle fontWeight={700}>Assign User to Service</DialogTitle>
      <Box component="form" onSubmit={handleSubmit}>
        <DialogContent sx={{ pt: 1 }}>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            Assign a team member to manage <strong>{serviceName}</strong>.
          </Typography>

          {isError && (
            <Alert severity="error" sx={{ mb: 2, borderRadius: 2 }}>
              {error?.message || "Failed to load users list"}
            </Alert>
          )}

          {loadingUsers ? (
            <Box sx={{ display: "flex", justifyContent: "center", py: 3 }}>
              <CircularProgress size={24} />
            </Box>
          ) : (
            <FormControl fullWidth margin="dense">
              <InputLabel id="assign-user-label">Select User</InputLabel>
              <Select
                labelId="assign-user-label"
                id="assign-user"
                value={selectedUserId}
                label="Select User"
                onChange={(e) => setSelectedUserId(e.target.value)}
                disabled={submitting}
              >
                {usersData?.map((u) => (
                  <MenuItem key={u.id} value={u.id}>
                    {u.email} ({u.role})
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          )}
        </DialogContent>

        <DialogActions sx={{ px: 3, pb: 2.5 }}>
          <Button onClick={handleClose} disabled={submitting} color="inherit">
            Cancel
          </Button>
          <Button
            type="submit"
            variant="contained"
            disabled={submitting || !selectedUserId}
            sx={{ px: 3 }}
          >
            {submitting ? (
              <CircularProgress size={20} color="inherit" />
            ) : (
              "Assign User"
            )}
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  );
}
