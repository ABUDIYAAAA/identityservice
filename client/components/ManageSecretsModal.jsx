"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Box from "@mui/material/Box";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TextField from "@mui/material/TextField";
import CircularProgress from "@mui/material/CircularProgress";
import Chip from "@mui/material/Chip";
import IconButton from "@mui/material/IconButton";
import Paper from "@mui/material/Paper";
import Alert from "@mui/material/Alert";
import { toast } from "sonner";
import api from "@/lib/api";
import SecretDisplayModal from "./SecretDisplayModal";

export default function ManageSecretsModal({ open, onClose, service }) {
  const queryClient = useQueryClient();
  const [newSecretName, setNewSecretName] = useState("");
  const [generating, setGenerating] = useState(false);
  const [newlyGeneratedSecret, setNewlyGeneratedSecret] = useState(null);

  const serviceId = service?.id;

  const {
    data: secretsList,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ["service-secrets", serviceId],
    queryFn: async () => {
      if (!serviceId) return [];
      const res = await api.get(`/services/${serviceId}/secrets`);
      return res.data?.data || [];
    },
    enabled: Boolean(open && serviceId),
  });

  const handleGenerateSecret = async (e) => {
    e.preventDefault();
    if (!newSecretName) {
      toast.error("Please enter a name for the secret");
      return;
    }

    setGenerating(true);
    try {
      const res = await api.post(`/services/${serviceId}/secrets`, {
        name: newSecretName,
      });
      const data = res.data?.data;
      toast.success("Client secret generated!");
      setNewSecretName("");
      refetch();
      if (data?.raw_secret) {
        setNewlyGeneratedSecret({
          rawSecret: data.raw_secret,
          clientId: service?.client_id,
        });
      }
    } catch (err) {
      toast.error(err.message || "Failed to generate secret");
    } finally {
      setGenerating(false);
    }
  };

  const deleteSecretMutation = useMutation({
    mutationFn: async (secretId) => {
      await api.delete(`/services/${serviceId}/secrets/${secretId}`);
    },
    onSuccess: () => {
      toast.success("Secret revoked successfully");
      refetch();
    },
    onError: (err) => {
      toast.error(err.message || "Failed to revoke secret");
    },
  });

  return (
    <>
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
        <DialogTitle fontWeight={700}>
          Manage API Secrets
          {service?.name && (
            <Typography variant="body2" color="text.secondary">
              Service: <strong>{service.name}</strong>
            </Typography>
          )}
        </DialogTitle>

        <DialogContent sx={{ pt: 1 }}>
          {/* Generate Secret Section */}
          <Box
            component="form"
            onSubmit={handleGenerateSecret}
            sx={{
              mb: 3,
              p: 2,
              backgroundColor: "#f8fafc",
              borderRadius: 2,
              border: "1px solid",
              borderColor: "divider",
            }}
          >
            <Typography variant="subtitle2" fontWeight={600} gutterBottom>
              Generate New Secret
            </Typography>
            <Box sx={{ display: "flex", gap: 1.5, mt: 1 }}>
              <TextField
                size="small"
                fullWidth
                placeholder="Secret identifier (e.g. Primary Key, Production)"
                value={newSecretName}
                onChange={(e) => setNewSecretName(e.target.value)}
                disabled={generating}
              />
              <Button
                type="submit"
                variant="contained"
                disabled={generating}
                sx={{ whiteSpace: "nowrap", px: 2.5 }}
              >
                {generating ? (
                  <CircularProgress size={20} color="inherit" />
                ) : (
                  "Generate"
                )}
              </Button>
            </Box>
          </Box>

          {isError && (
            <Alert severity="error" sx={{ mb: 2, borderRadius: 2 }}>
              {error?.message || "Failed to load secrets"}
            </Alert>
          )}

          {/* Secrets List Table */}
          <TableContainer
            component={Paper}
            elevation={0}
            sx={{
              border: "1px solid",
              borderColor: "divider",
              borderRadius: 2,
            }}
          >
            <Table size="small">
              <TableHead sx={{ backgroundColor: "#f8fafc" }}>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>Name</TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>Key Prefix</TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>Created</TableCell>
                  <TableCell align="right" sx={{ fontWeight: 600 }}>
                    Actions
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {isLoading ? (
                  <TableRow>
                    <TableCell colSpan={4} align="center" sx={{ py: 3 }}>
                      <CircularProgress size={24} />
                    </TableCell>
                  </TableRow>
                ) : secretsList?.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={4}
                      align="center"
                      sx={{ py: 3, color: "text.secondary" }}
                    >
                      No active secrets found.
                    </TableCell>
                  </TableRow>
                ) : (
                  secretsList?.map((sec) => (
                    <TableRow key={sec.id}>
                      <TableCell sx={{ fontWeight: 500 }}>{sec.name}</TableCell>
                      <TableCell>
                        <Chip
                          label={`${sec.secret_prefix || "id_sec"}...`}
                          size="small"
                          sx={{
                            fontFamily: "monospace",
                            fontSize: "0.75rem",
                            backgroundColor: "#f1f5f9",
                          }}
                        />
                      </TableCell>
                      <TableCell
                        sx={{ fontSize: "0.8rem", color: "text.secondary" }}
                      >
                        {sec.created_at
                          ? new Date(sec.created_at).toLocaleDateString()
                          : "-"}
                      </TableCell>
                      <TableCell align="right">
                        <Button
                          size="small"
                          color="error"
                          onClick={() => deleteSecretMutation.mutate(sec.id)}
                          disabled={deleteSecretMutation.isPending}
                          sx={{
                            textTransform: "none",
                            fontWeight: 600,
                            fontSize: "0.75rem",
                          }}
                        >
                          Revoke
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </DialogContent>

        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={onClose} variant="outlined" color="inherit">
            Close
          </Button>
        </DialogActions>
      </Dialog>

      {/* Show newly generated raw secret if available */}
      {newlyGeneratedSecret && (
        <SecretDisplayModal
          open={Boolean(newlyGeneratedSecret)}
          onClose={() => setNewlyGeneratedSecret(null)}
          rawSecret={newlyGeneratedSecret.rawSecret}
          clientId={newlyGeneratedSecret.clientId}
          serviceName={service?.name}
        />
      )}
    </>
  );
}
