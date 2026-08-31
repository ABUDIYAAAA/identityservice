"use client";

import { useState, use } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import Button from "@mui/material/Button";
import Chip from "@mui/material/Chip";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import CircularProgress from "@mui/material/CircularProgress";
import Alert from "@mui/material/Alert";
import Tooltip from "@mui/material/Tooltip";
import TextField from "@mui/material/TextField";
import ProtectedRoute from "@/components/ProtectedRoute";
import DashboardLayout from "@/components/DashboardLayout";
import SecretDisplayModal from "@/components/SecretDisplayModal";
import AssignUserModal from "@/components/AssignUserModal";
import AddPermissionModal from "@/components/AddPermissionModal";
import ConfirmDeleteDialog from "@/components/ConfirmDeleteDialog";
import { useAuth } from "@/hooks/useAuth";
import { toast } from "sonner";
import api from "@/lib/api";

function TabPanel({ children, value, index }) {
  return (
    <div role="tabpanel" hidden={value !== index}>
      {value === index && <Box sx={{ pt: 3 }}>{children}</Box>}
    </div>
  );
}

function ServiceDetailsContent({ serviceId }) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { user: currentUser } = useAuth();

  const [activeTab, setActiveTab] = useState(0);
  const [newSecretName, setNewSecretName] = useState("");
  const [generatingSecret, setGeneratingSecret] = useState(false);
  const [newlyGeneratedSecret, setNewlyGeneratedSecret] = useState(null);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const [assignUserModalOpen, setAssignUserModalOpen] = useState(false);
  const [addPermModalOpen, setAddPermModalOpen] = useState(false);

  const isAdmin = currentUser?.role === "admin";

  // 1. Fetch Service Details
  const {
    data: service,
    isLoading: loadingService,
    isError: serviceError,
    error: serviceErrObj,
  } = useQuery({
    queryKey: ["service-details", serviceId],
    queryFn: async () => {
      const res = await api.get(`/services/${serviceId}`);
      return res.data?.data;
    },
  });

  // 2. Fetch Service Secrets
  const {
    data: secretsList,
    isLoading: loadingSecrets,
    refetch: refetchSecrets,
  } = useQuery({
    queryKey: ["service-secrets", serviceId],
    queryFn: async () => {
      const res = await api.get(`/services/${serviceId}/secrets`);
      return res.data?.data || [];
    },
    enabled: Boolean(serviceId),
  });

  // 3. Fetch Service Permissions (Allowed Target Services)
  const {
    data: permissionsList,
    isLoading: loadingPermissions,
    refetch: refetchPermissions,
  } = useQuery({
    queryKey: ["service-permissions", serviceId],
    queryFn: async () => {
      const res = await api.get(`/services/${serviceId}/permissions`);
      return res.data?.data || [];
    },
    enabled: Boolean(serviceId),
  });

  // Status Toggle Mutation
  const updateStatusMutation = useMutation({
    mutationFn: async (newStatus) => {
      await api.patch(`/services/${serviceId}/status`, {
        is_active: newStatus,
      });
    },
    onSuccess: (_, newStatus) => {
      toast.success(`Service ${newStatus ? "activated" : "deactivated"}`);
      queryClient.invalidateQueries(["service-details", serviceId]);
      queryClient.invalidateQueries(["services"]);
    },
    onError: (err) => {
      toast.error(err.message || "Failed to update service status");
    },
  });

  // Service Delete Mutation
  const handleDeleteConfirm = async () => {
    setDeleting(true);
    try {
      await api.delete(`/services/${serviceId}`);
      toast.success("Service deleted successfully");
      queryClient.invalidateQueries(["services"]);
      router.push("/services");
    } catch (err) {
      toast.error(err.message || "Failed to delete service");
    } finally {
      setDeleting(false);
    }
  };

  // Generate Secret Handler
  const handleGenerateSecret = async (e) => {
    e.preventDefault();
    if (!newSecretName) {
      toast.error("Please enter a name for the secret");
      return;
    }

    setGeneratingSecret(true);
    try {
      const res = await api.post(`/services/${serviceId}/secrets`, {
        name: newSecretName,
      });
      const data = res.data?.data;
      toast.success("Client secret generated!");
      setNewSecretName("");
      refetchSecrets();
      if (data?.raw_secret) {
        setNewlyGeneratedSecret({
          rawSecret: data.raw_secret,
          clientId: service?.client_id,
        });
      }
    } catch (err) {
      toast.error(err.message || "Failed to generate secret");
    } finally {
      setGeneratingSecret(false);
    }
  };

  // Revoke Secret Handler
  const handleRevokeSecret = async (secretId) => {
    try {
      await api.delete(`/services/${serviceId}/secrets/${secretId}`);
      toast.success("Secret revoked successfully");
      refetchSecrets();
    } catch (err) {
      toast.error(err.message || "Failed to revoke secret");
    }
  };

  // Revoke Permission Link Handler
  const handleRevokePermission = async (targetId) => {
    try {
      await api.delete(`/services/${serviceId}/permissions/${targetId}`);
      toast.success("Permission link revoked");
      refetchPermissions();
    } catch (err) {
      toast.error(err.message || "Failed to revoke permission link");
    }
  };

  const handleCopyClientId = () => {
    if (!service?.client_id) return;
    navigator.clipboard.writeText(service.client_id);
    toast.success("Client ID copied to clipboard");
  };

  if (loadingService) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", py: 8 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (serviceError || !service) {
    return (
      <Box>
        <Button
          component={Link}
          href="/services"
          color="inherit"
          sx={{ mb: 3 }}
        >
          &larr; Back to Services
        </Button>
        <Alert severity="error" sx={{ borderRadius: 2 }}>
          {serviceErrObj?.message || "Service not found or access denied."}
        </Alert>
      </Box>
    );
  }

  return (
    <Box>
      {/* Back Button */}
      <Button
        component={Link}
        href="/services"
        color="inherit"
        sx={{ mb: 3, textTransform: "none", fontWeight: 600 }}
      >
        &larr; Back to Services
      </Button>

      {/* Service Header Card */}
      <Paper
        elevation={0}
        sx={{
          p: 4,
          borderRadius: 3,
          border: "1px solid",
          borderColor: "divider",
          backgroundColor: "#ffffff",
          mb: 4,
        }}
      >
        <Box
          sx={{
            display: "flex",
            alignItems: "flex-start",
            justifyContent: "space-between",
          }}
        >
          <Box>
            <Box
              sx={{ display: "flex", alignItems: "center", gap: 1.5, mb: 1 }}
            >
              <Typography variant="h4" fontWeight={700}>
                {service.name}
              </Typography>
              <Chip
                label={service.is_active ? "Active" : "Inactive"}
                size="small"
                sx={{
                  fontWeight: 600,
                  fontSize: "0.75rem",
                  backgroundColor: service.is_active ? "#dcfce7" : "#fee2e2",
                  color: service.is_active ? "#15803d" : "#b91c1c",
                }}
              />
            </Box>

            {service.description && (
              <Typography variant="body1" color="text.secondary" sx={{ mb: 2 }}>
                {service.description}
              </Typography>
            )}

            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Typography variant="caption" color="text.secondary">
                Client ID:
              </Typography>
              <Tooltip title="Click to copy Client ID" placement="top">
                <Chip
                  label={service.client_id}
                  onClick={handleCopyClientId}
                  size="small"
                  clickable
                  sx={{
                    fontFamily: "monospace",
                    fontSize: "0.8rem",
                    fontWeight: 600,
                    backgroundColor: "#f1f5f9",
                  }}
                />
              </Tooltip>
            </Box>
          </Box>

          {isAdmin && (
            <Box sx={{ display: "flex", gap: 1.5 }}>
              <Button
                variant="outlined"
                color={service.is_active ? "warning" : "success"}
                onClick={() => updateStatusMutation.mutate(!service.is_active)}
                disabled={updateStatusMutation.isPending}
                sx={{ textTransform: "none", fontWeight: 600 }}
              >
                {service.is_active ? "Deactivate" : "Activate"}
              </Button>
              <Button
                variant="contained"
                color="error"
                onClick={() => setDeleteOpen(true)}
                sx={{ textTransform: "none", fontWeight: 600 }}
              >
                Delete Service
              </Button>
            </Box>
          )}
        </Box>
      </Paper>

      {/* Tabs Container */}
      <Paper
        elevation={0}
        sx={{
          borderRadius: 3,
          border: "1px solid",
          borderColor: "divider",
          backgroundColor: "#ffffff",
          p: 3,
        }}
      >
        <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
          <Tabs
            value={activeTab}
            onChange={(e, val) => setActiveTab(val)}
            textColor="primary"
            indicatorColor="primary"
          >
            <Tab
              label="API Secrets & Keys"
              sx={{ fontWeight: 600, textTransform: "none" }}
            />
            <Tab
              label="Inter-Service Permissions"
              sx={{ fontWeight: 600, textTransform: "none" }}
            />
            <Tab
              label="User Assignments"
              sx={{ fontWeight: 600, textTransform: "none" }}
            />
          </Tabs>
        </Box>

        {/* TAB 0: API Secrets */}
        <TabPanel value={activeTab} index={0}>
          <Box
            component="form"
            onSubmit={handleGenerateSecret}
            sx={{
              mb: 3,
              p: 2.5,
              backgroundColor: "#f8fafc",
              borderRadius: 2,
              border: "1px solid",
              borderColor: "divider",
            }}
          >
            <Typography variant="subtitle2" fontWeight={600} gutterBottom>
              Generate New Client Secret
            </Typography>
            <Box sx={{ display: "flex", gap: 1.5, mt: 1 }}>
              <TextField
                size="small"
                fullWidth
                placeholder="Secret name (e.g., Primary Secret, Staging Key)"
                value={newSecretName}
                onChange={(e) => setNewSecretName(e.target.value)}
                disabled={generatingSecret}
              />
              <Button
                type="submit"
                variant="contained"
                disabled={generatingSecret}
                sx={{ whiteSpace: "nowrap", px: 3 }}
              >
                {generatingSecret ? (
                  <CircularProgress size={20} color="inherit" />
                ) : (
                  "Generate Secret"
                )}
              </Button>
            </Box>
          </Box>

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
                  <TableCell sx={{ fontWeight: 600 }}>Prefix</TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>Created At</TableCell>
                  <TableCell align="right" sx={{ fontWeight: 600 }}>
                    Actions
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {loadingSecrets ? (
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
                      sx={{ py: 4, color: "text.secondary" }}
                    >
                      No active secrets found.
                    </TableCell>
                  </TableRow>
                ) : (
                  secretsList?.map((sec) => (
                    <TableRow key={sec.id}>
                      <TableCell sx={{ fontWeight: 600 }}>{sec.name}</TableCell>
                      <TableCell>
                        <Chip
                          label={`${sec.secret_prefix || "sec_live"}...`}
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
                          onClick={() => handleRevokeSecret(sec.id)}
                          sx={{ textTransform: "none", fontWeight: 600 }}
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
        </TabPanel>

        {/* TAB 1: Inter-Service Link Permissions */}
        <TabPanel value={activeTab} index={1}>
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              mb: 3,
            }}
          >
            <Typography variant="body2" color="text.secondary">
              Configure target microservices that{" "}
              <strong>{service.name}</strong> is authorized to call via M2M
              verification.
            </Typography>

            {isAdmin && (
              <Button
                variant="contained"
                disableElevation
                size="small"
                onClick={() => setAddPermModalOpen(true)}
                sx={{ px: 2.5 }}
              >
                + Add Target Link
              </Button>
            )}
          </Box>

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
                  <TableCell sx={{ fontWeight: 600 }}>Target Service</TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>
                    Target Client ID
                  </TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>Status</TableCell>
                  <TableCell align="right" sx={{ fontWeight: 600 }}>
                    Actions
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {loadingPermissions ? (
                  <TableRow>
                    <TableCell colSpan={4} align="center" sx={{ py: 3 }}>
                      <CircularProgress size={24} />
                    </TableCell>
                  </TableRow>
                ) : permissionsList?.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={4}
                      align="center"
                      sx={{ py: 4, color: "text.secondary" }}
                    >
                      No target service permissions configured.
                    </TableCell>
                  </TableRow>
                ) : (
                  permissionsList?.map((tgt) => (
                    <TableRow key={tgt.id}>
                      <TableCell sx={{ fontWeight: 600 }}>{tgt.name}</TableCell>
                      <TableCell>
                        <Chip
                          label={tgt.client_id}
                          size="small"
                          sx={{
                            fontFamily: "monospace",
                            fontSize: "0.75rem",
                            backgroundColor: "#f1f5f9",
                          }}
                        />
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={tgt.is_active ? "Active" : "Inactive"}
                          size="small"
                          sx={{
                            fontWeight: 600,
                            fontSize: "0.7rem",
                            backgroundColor: tgt.is_active
                              ? "#dcfce7"
                              : "#fee2e2",
                            color: tgt.is_active ? "#15803d" : "#b91c1c",
                          }}
                        />
                      </TableCell>
                      <TableCell align="right">
                        {isAdmin && (
                          <Button
                            size="small"
                            color="error"
                            onClick={() => handleRevokePermission(tgt.id)}
                            sx={{ textTransform: "none", fontWeight: 600 }}
                          >
                            Remove Link
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </TabPanel>

        {/* TAB 2: User Assignments */}
        <TabPanel value={activeTab} index={2}>
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              mb: 3,
            }}
          >
            <Typography variant="body2" color="text.secondary">
              Team members assigned to manage <strong>{service.name}</strong>.
            </Typography>

            {isAdmin && (
              <Button
                variant="contained"
                disableElevation
                size="small"
                onClick={() => setAssignUserModalOpen(true)}
                sx={{ px: 2.5 }}
              >
                + Assign User
              </Button>
            )}
          </Box>

          <Alert severity="info" sx={{ borderRadius: 2 }}>
            Administrators automatically have full management access to all
            services. Assigned users gain delegate permissions.
          </Alert>
        </TabPanel>
      </Paper>

      {/* Modals */}
      {newlyGeneratedSecret && (
        <SecretDisplayModal
          open={Boolean(newlyGeneratedSecret)}
          onClose={() => setNewlyGeneratedSecret(null)}
          rawSecret={newlyGeneratedSecret.rawSecret}
          clientId={newlyGeneratedSecret.clientId}
          serviceName={service.name}
        />
      )}

      <AssignUserModal
        open={assignUserModalOpen}
        onClose={() => setAssignUserModalOpen(false)}
        serviceId={serviceId}
        serviceName={service.name}
      />

      <AddPermissionModal
        open={addPermModalOpen}
        onClose={() => setAddPermModalOpen(false)}
        serviceId={serviceId}
        serviceName={service.name}
        onSuccess={refetchPermissions}
      />

      <ConfirmDeleteDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={handleDeleteConfirm}
        title="Delete Service"
        targetName={service.name}
        loading={deleting}
      />
    </Box>
  );
}

export default function ServiceDetailsPage({ params: paramsPromise }) {
  const params = use(paramsPromise);
  return (
    <ProtectedRoute>
      <DashboardLayout>
        <ServiceDetailsContent serviceId={params.id} />
      </DashboardLayout>
    </ProtectedRoute>
  );
}
