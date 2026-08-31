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

export default function AddPermissionModal({
  open,
  onClose,
  serviceId,
  serviceName,
  onSuccess,
}) {
  const [selectedTargetId, setSelectedTargetId] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const {
    data: servicesData,
    isLoading: loadingServices,
    isError,
    error,
  } = useQuery({
    queryKey: ["all-services-for-permission"],
    queryFn: async () => {
      const res = await api.get("/services", { params: { limit: 100 } });
      return res.data?.data?.services || [];
    },
    enabled: Boolean(open),
  });

  const targetOptions = servicesData?.filter((s) => s.id !== serviceId) || [];

  const handleClose = () => {
    if (submitting) return;
    setSelectedTargetId("");
    onClose();
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!selectedTargetId) {
      toast.error("Please select a target service");
      return;
    }

    setSubmitting(true);
    try {
      await api.post(`/services/${serviceId}/permissions`, {
        target_service_id: selectedTargetId,
      });
      toast.success("Inter-service permission granted successfully");
      setSelectedTargetId("");
      onSuccess?.();
      onClose();
    } catch (err) {
      toast.error(err.message || "Failed to grant permission");
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
      <DialogTitle fontWeight={700}>Add Target Service Link</DialogTitle>
      <Box component="form" onSubmit={handleSubmit}>
        <DialogContent sx={{ pt: 1 }}>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            Allow <strong>{serviceName}</strong> to issue M2M tokens and make
            authenticated requests to a target service.
          </Typography>

          {isError && (
            <Alert severity="error" sx={{ mb: 2, borderRadius: 2 }}>
              {error?.message || "Failed to load services list"}
            </Alert>
          )}

          {loadingServices ? (
            <Box sx={{ display: "flex", justifyContent: "center", py: 3 }}>
              <CircularProgress size={24} />
            </Box>
          ) : targetOptions.length === 0 ? (
            <Alert severity="info" sx={{ borderRadius: 2 }}>
              No other target services available to link.
            </Alert>
          ) : (
            <FormControl fullWidth margin="dense">
              <InputLabel id="permission-target-label">
                Target Service
              </InputLabel>
              <Select
                labelId="permission-target-label"
                id="permission-target"
                value={selectedTargetId}
                label="Target Service"
                onChange={(e) => setSelectedTargetId(e.target.value)}
                disabled={submitting}
              >
                {targetOptions.map((s) => (
                  <MenuItem key={s.id} value={s.id}>
                    {s.name} ({s.client_id})
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
            disabled={submitting || !selectedTargetId}
            sx={{ px: 3 }}
          >
            {submitting ? (
              <CircularProgress size={20} color="inherit" />
            ) : (
              "Grant Access"
            )}
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  );
}
