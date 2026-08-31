"use client";

import { useState } from "react";
import Link from "next/link";
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
import Avatar from "@mui/material/Avatar";
import Chip from "@mui/material/Chip";
import IconButton from "@mui/material/IconButton";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import Skeleton from "@mui/material/Skeleton";
import Alert from "@mui/material/Alert";
import ProtectedRoute from "@/components/ProtectedRoute";
import DashboardLayout from "@/components/DashboardLayout";
import InviteUserModal from "@/components/InviteUserModal";
import ConfirmDeleteDialog from "@/components/ConfirmDeleteDialog";
import { useAuth } from "@/hooks/useAuth";
import { toast } from "sonner";
import api from "@/lib/api";

function UserRow({ user: targetUser, currentUser, onRefresh }) {
  const queryClient = useQueryClient();
  const [anchorEl, setAnchorEl] = useState(null);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const openMenu = Boolean(anchorEl);

  const handleOpenMenu = (e) => setAnchorEl(e.currentTarget);
  const handleCloseMenu = () => setAnchorEl(null);

  const isSelf = targetUser.id === currentUser?.id;
  const isAdmin = currentUser?.role === "admin";

  // Role Update Mutation
  const updateRoleMutation = useMutation({
    mutationFn: async (newRole) => {
      await api.patch(`/users/${targetUser.id}/role`, { role: newRole });
    },
    onSuccess: (_, newRole) => {
      toast.success(`Role updated to ${newRole}`);
      queryClient.invalidateQueries(["users"]);
    },
    onError: (err) => {
      toast.error(err.message || "Failed to update role");
    },
  });

  // Status Toggle Mutation
  const updateStatusMutation = useMutation({
    mutationFn: async (newStatus) => {
      await api.patch(`/users/${targetUser.id}/status`, {
        is_active: newStatus,
      });
    },
    onSuccess: (_, newStatus) => {
      toast.success(`User ${newStatus ? "activated" : "deactivated"}`);
      queryClient.invalidateQueries(["users"]);
    },
    onError: (err) => {
      toast.error(err.message || "Failed to update status");
    },
  });

  // User Delete Mutation
  const handleDeleteConfirm = async () => {
    setDeleting(true);
    try {
      await api.delete(`/users/${targetUser.id}`);
      toast.success("User deleted successfully");
      queryClient.invalidateQueries(["users"]);
      setDeleteOpen(false);
    } catch (err) {
      toast.error(err.message || "Failed to delete user");
    } finally {
      setDeleting(false);
    }
  };

  const userInitial = targetUser.email
    ? targetUser.email.charAt(0).toUpperCase()
    : "U";
  const formattedDate = targetUser.created_at
    ? new Date(targetUser.created_at).toLocaleDateString("en-US", {
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
          <Box sx={{ display: "flex", alignItems: "center", gap: 1.5 }}>
            <Avatar
              sx={{
                width: 34,
                height: 34,
                backgroundColor: "primary.main",
                fontSize: "0.85rem",
                fontWeight: 600,
              }}
            >
              {userInitial}
            </Avatar>
            <Box>
              <Typography variant="body2" fontWeight={600} color="text.primary">
                {targetUser.email}
              </Typography>
              {isSelf && (
                <Typography
                  variant="caption"
                  color="primary.main"
                  fontWeight={500}
                >
                  (You)
                </Typography>
              )}
            </Box>
          </Box>
        </TableCell>

        <TableCell>
          <Chip
            label={targetUser.role}
            size="small"
            sx={{
              fontWeight: 600,
              fontSize: "0.725rem",
              textTransform: "uppercase",
              backgroundColor:
                targetUser.role === "admin" ? "#0f172a" : "#f1f5f9",
              color: targetUser.role === "admin" ? "#ffffff" : "#475569",
            }}
          />
        </TableCell>

        <TableCell>
          <Chip
            label={targetUser.is_active ? "Active" : "Inactive"}
            size="small"
            sx={{
              fontWeight: 600,
              fontSize: "0.725rem",
              backgroundColor: targetUser.is_active ? "#dcfce7" : "#fee2e2",
              color: targetUser.is_active ? "#15803d" : "#b91c1c",
            }}
          />
        </TableCell>

        <TableCell>
          <Typography variant="body2" color="text.secondary">
            {formattedDate}
          </Typography>
        </TableCell>

        <TableCell align="right">
          {isAdmin && !isSelf ? (
            <>
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
                {targetUser.role === "user" ? (
                  <MenuItem
                    onClick={() => {
                      handleCloseMenu();
                      updateRoleMutation.mutate("admin");
                    }}
                    sx={{ fontSize: "0.85rem", py: 1 }}
                  >
                    Make Admin
                  </MenuItem>
                ) : (
                  <MenuItem
                    onClick={() => {
                      handleCloseMenu();
                      updateRoleMutation.mutate("user");
                    }}
                    sx={{ fontSize: "0.85rem", py: 1 }}
                  >
                    Make User
                  </MenuItem>
                )}

                <MenuItem
                  onClick={() => {
                    handleCloseMenu();
                    updateStatusMutation.mutate(!targetUser.is_active);
                  }}
                  sx={{ fontSize: "0.85rem", py: 1 }}
                >
                  {targetUser.is_active ? "Deactivate" : "Activate"}
                </MenuItem>

                <MenuItem
                  onClick={() => {
                    handleCloseMenu();
                    setDeleteOpen(true);
                  }}
                  sx={{ fontSize: "0.85rem", color: "error.main", py: 1 }}
                >
                  Delete User
                </MenuItem>
              </Menu>
            </>
          ) : (
            <Typography variant="caption" color="text.disabled">
              {isSelf ? "Current User" : "Read-only"}
            </Typography>
          )}
        </TableCell>
      </TableRow>

      <ConfirmDeleteDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={handleDeleteConfirm}
        title="Delete User"
        targetName={targetUser.email}
        loading={deleting}
      />
    </>
  );
}

function UsersContent() {
  const { user: currentUser } = useAuth();
  const queryClient = useQueryClient();

  const [page, setPage] = useState(0); // MUI TablePagination is 0-indexed
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [inviteModalOpen, setInviteModalOpen] = useState(false);

  const apiPage = page + 1; // Backend page is 1-indexed

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["users", apiPage, rowsPerPage],
    queryFn: async () => {
      const res = await api.get("/users", {
        params: { page: apiPage, limit: rowsPerPage },
      });
      return res.data?.data;
    },
  });

  const usersList = data?.users || [];
  const totalCount = data?.total_count || 0;

  const handleChangePage = (event, newPage) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const isAdmin = currentUser?.role === "admin";

  if (!isAdmin) {
    return (
      <Paper
        elevation={0}
        sx={{
          p: 4,
          borderRadius: 3,
          border: "1px solid",
          borderColor: "divider",
          textAlign: "center",
        }}
      >
        <Alert severity="error" sx={{ mb: 2, borderRadius: 2 }}>
          Access Restricted: User management is restricted to administrators
          only.
        </Alert>
        <Button component={Link} href="/" variant="outlined" sx={{ mt: 1 }}>
          Return to Dashboard
        </Button>
      </Paper>
    );
  }

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
            User Management
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage your organization's team members, roles, and pending
            invitations
          </Typography>
        </Box>

        {isAdmin && (
          <Button
            variant="contained"
            disableElevation
            onClick={() => setInviteModalOpen(true)}
            sx={{ px: 3, py: 1 }}
          >
            + Invite User
          </Button>
        )}
      </Box>

      {isError && (
        <Alert severity="error" sx={{ mb: 3, borderRadius: 2 }}>
          {error?.message || "Failed to load users list"}
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
                  User
                </TableCell>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Role
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
                      <Box
                        sx={{ display: "flex", alignItems: "center", gap: 1.5 }}
                      >
                        <Skeleton variant="circular" width={34} height={34} />
                        <Skeleton variant="text" width={160} height={24} />
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="rounded" width={50} height={24} />
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
              ) : usersList.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} align="center" sx={{ py: 6 }}>
                    <Typography variant="body2" color="text.secondary">
                      No users found.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                usersList.map((usr) => (
                  <UserRow
                    key={usr.id}
                    user={usr}
                    currentUser={currentUser}
                    onRefresh={refetch}
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

      <InviteUserModal
        open={inviteModalOpen}
        onClose={() => setInviteModalOpen(false)}
        onSuccess={() => queryClient.invalidateQueries(["users"])}
      />
    </Box>
  );
}

export default function UsersPage() {
  return (
    <ProtectedRoute>
      <DashboardLayout>
        <UsersContent />
      </DashboardLayout>
    </ProtectedRoute>
  );
}
