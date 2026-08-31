"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import Button from "@mui/material/Button";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TablePagination from "@mui/material/TablePagination";
import Chip from "@mui/material/Chip";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import Skeleton from "@mui/material/Skeleton";
import Alert from "@mui/material/Alert";
import Tooltip from "@mui/material/Tooltip";
import ProtectedRoute from "@/components/ProtectedRoute";
import DashboardLayout from "@/components/DashboardLayout";
import CreateServiceModal from "@/components/CreateServiceModal";
import ManageSecretsModal from "@/components/ManageSecretsModal";
import SecretDisplayModal from "@/components/SecretDisplayModal";
import ConfirmDeleteDialog from "@/components/ConfirmDeleteDialog";
import { useAuth } from "@/hooks/useAuth";
import { toast } from "sonner";
import api from "@/lib/api";

function ServiceRow({ service, currentUser, onManageSecrets }) {
  const queryClient = useQueryClient();
  const [anchorEl, setAnchorEl] = useState(null);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const openMenu = Boolean(anchorEl);

  const handleOpenMenu = (e) => setAnchorEl(e.currentTarget);
  const handleCloseMenu = () => setAnchorEl(null);

  const isAdmin = currentUser?.role === "admin";

  // Status Toggle Mutation
  const updateStatusMutation = useMutation({
    mutationFn: async (newStatus) => {
      await api.patch(`/services/${service.id}/status`, {
        is_active: newStatus,
      });
    },
    onSuccess: (_, newStatus) => {
      toast.success(`Service ${newStatus ? "activated" : "deactivated"}`);
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
      await api.delete(`/services/${service.id}`);
      toast.success("Service deleted successfully");
      queryClient.invalidateQueries(["services"]);
      setDeleteOpen(false);
    } catch (err) {
      toast.error(err.message || "Failed to delete service");
    } finally {
      setDeleting(false);
    }
  };

  const handleCopyClientId = () => {
    navigator.clipboard.writeText(service.client_id);
    toast.success("Client ID copied to clipboard");
  };

  const formattedDate = service.created_at
    ? new Date(service.created_at).toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year: "numeric",
      })
    : "-";

  return (
    <>
      <TableRow
        hover
        sx={{ "&:last-child td, &:last-child th": { border: 0 } }}
      >
        <TableCell>
          <Box>
            <Typography variant="body2" fontWeight={600} color="text.primary">
              {service.name}
            </Typography>
            {service.description && (
              <Typography
                variant="caption"
                color="text.secondary"
                display="block"
                noWrap
                sx={{ maxWidth: 300 }}
              >
                {service.description}
              </Typography>
            )}
          </Box>
        </TableCell>

        <TableCell>
          <Tooltip title="Click to copy Client ID" placement="top">
            <Chip
              label={service.client_id}
              onClick={handleCopyClientId}
              size="small"
              clickable
              sx={{
                fontFamily: "monospace",
                fontSize: "0.75rem",
                backgroundColor: "#f1f5f9",
                borderColor: "#e2e8f0",
                "&:hover": {
                  backgroundColor: "#e2e8f0",
                },
              }}
            />
          </Tooltip>
        </TableCell>

        <TableCell>
          <Chip
            label={service.is_active ? "Active" : "Inactive"}
            size="small"
            sx={{
              fontWeight: 600,
              fontSize: "0.725rem",
              backgroundColor: service.is_active ? "#dcfce7" : "#fee2e2",
              color: service.is_active ? "#15803d" : "#b91c1c",
            }}
          />
        </TableCell>

        <TableCell>
          <Typography variant="body2" color="text.secondary">
            {formattedDate}
          </Typography>
        </TableCell>

        <TableCell align="right">
          <Button
            size="small"
            variant="outlined"
            color="inherit"
            onClick={handleOpenMenu}
            sx={{
              borderRadius: 2,
              textTransform: "none",
              fontWeight: 600,
              fontSize: "0.8rem",
              py: 0.4,
              borderColor: "divider",
            }}
          >
            Manage
          </Button>

          <Menu
            anchorEl={anchorEl}
            open={openMenu}
            onClose={handleCloseMenu}
            slotProps={{
              paper: {
                elevation: 0,
                sx: {
                  width: 170,
                  borderRadius: 2.5,
                  border: "1px solid",
                  borderColor: "divider",
                  boxShadow: "0 4px 12px 0 rgb(0 0 0 / 0.08)",
                  py: 0.5,
                },
              },
            }}
          >
            <MenuItem
              onClick={() => {
                handleCloseMenu();
                onManageSecrets(service);
              }}
              sx={{ fontSize: "0.85rem", py: 1 }}
            >
              Client Secrets
            </MenuItem>

            {isAdmin && (
              <MenuItem
                onClick={() => {
                  handleCloseMenu();
                  updateStatusMutation.mutate(!service.is_active);
                }}
                sx={{ fontSize: "0.85rem", py: 1 }}
              >
                {service.is_active ? "Deactivate" : "Activate"}
              </MenuItem>
            )}

            {isAdmin && (
              <MenuItem
                onClick={() => {
                  handleCloseMenu();
                  setDeleteOpen(true);
                }}
                sx={{ fontSize: "0.85rem", color: "error.main", py: 1 }}
              >
                Delete Service
              </MenuItem>
            )}
          </Menu>
        </TableCell>
      </TableRow>

      <ConfirmDeleteDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={handleDeleteConfirm}
        title="Delete Service"
        targetName={service.name}
        loading={deleting}
      />
    </>
  );
}

function ServicesContent() {
  const { user: currentUser } = useAuth();
  const queryClient = useQueryClient();

  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [secretsModalService, setSecretsModalService] = useState(null);
  const [initialSecretData, setInitialSecretData] = useState(null);

  const apiPage = page + 1;

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["services", apiPage, rowsPerPage],
    queryFn: async () => {
      const res = await api.get("/services", {
        params: { page: apiPage, limit: rowsPerPage },
      });
      return res.data?.data;
    },
  });

  const servicesList = data?.services || [];
  const totalCount = data?.total_count || 0;

  const handleChangePage = (event, newPage) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleServiceCreated = (createdData) => {
    queryClient.invalidateQueries(["services"]);
    if (createdData?.initial_secret) {
      setInitialSecretData({
        rawSecret: createdData.initial_secret,
        clientId: createdData.service?.client_id,
        serviceName: createdData.service?.name,
      });
    }
  };

  const isAdmin = currentUser?.role === "admin";

  return (
    <Box>
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          mb: 4,
        }}
      >
        <Box>
          <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
            Services & Applications
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage your registered microservices, client IDs, and secrets
          </Typography>
        </Box>

        {isAdmin && (
          <Button
            variant="contained"
            disableElevation
            onClick={() => setCreateModalOpen(true)}
            sx={{ px: 3, py: 1 }}
          >
            + Create Service
          </Button>
        )}
      </Box>

      {isError && (
        <Alert severity="error" sx={{ mb: 3, borderRadius: 2 }}>
          {error?.message || "Failed to load services list"}
        </Alert>
      )}

      <Paper
        elevation={0}
        sx={{
          borderRadius: 3,
          border: "1px solid",
          borderColor: "divider",
          overflow: "hidden",
          backgroundColor: "#ffffff",
        }}
      >
        <TableContainer>
          <Table sx={{ minWidth: 650 }}>
            <TableHead sx={{ backgroundColor: "#f8fafc" }}>
              <TableRow>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Service
                </TableCell>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Client ID
                </TableCell>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Status
                </TableCell>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Created At
                </TableCell>
                <TableCell
                  align="right"
                  sx={{ fontWeight: 600, color: "text.secondary" }}
                >
                  Actions
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, idx) => (
                  <TableRow key={idx}>
                    <TableCell>
                      <Skeleton variant="text" width={140} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="rounded" width={180} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="rounded" width={60} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" width={90} height={24} />
                    </TableCell>
                    <TableCell align="right">
                      <Skeleton
                        variant="rounded"
                        width={70}
                        height={28}
                        sx={{ ml: "auto" }}
                      />
                    </TableCell>
                  </TableRow>
                ))
              ) : servicesList.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} align="center" sx={{ py: 6 }}>
                    <Typography variant="body2" color="text.secondary">
                      No services found.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                servicesList.map((svc) => (
                  <ServiceRow
                    key={svc.id}
                    service={svc}
                    currentUser={currentUser}
                    onManageSecrets={(selectedSvc) =>
                      setSecretsModalService(selectedSvc)
                    }
                  />
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>

        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 50]}
          component="div"
          count={totalCount}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
          sx={{ borderTop: "1px solid", borderColor: "divider" }}
        />
      </Paper>

      {/* Modal Dialogs */}
      <CreateServiceModal
        open={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        onSuccess={handleServiceCreated}
      />

      <ManageSecretsModal
        open={Boolean(secretsModalService)}
        onClose={() => setSecretsModalService(null)}
        service={secretsModalService}
      />

      {initialSecretData && (
        <SecretDisplayModal
          open={Boolean(initialSecretData)}
          onClose={() => setInitialSecretData(null)}
          rawSecret={initialSecretData.rawSecret}
          clientId={initialSecretData.clientId}
          serviceName={initialSecretData.serviceName}
        />
      )}
    </Box>
  );
}

export default function ServicesPage() {
  return (
    <ProtectedRoute>
      <DashboardLayout>
        <ServicesContent />
      </DashboardLayout>
    </ProtectedRoute>
  );
}
