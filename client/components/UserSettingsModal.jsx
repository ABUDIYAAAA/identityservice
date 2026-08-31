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
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import TextField from "@mui/material/TextField";
import CircularProgress from "@mui/material/CircularProgress";
import Chip from "@mui/material/Chip";
import Paper from "@mui/material/Paper";
import Alert from "@mui/material/Alert";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import { toast } from "sonner";
import api from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";

function TabPanel({ children, value, index }) {
  return (
    <div role="tabpanel" hidden={value !== index}>
      {value === index && <Box sx={{ pt: 3 }}>{children}</Box>}
    </div>
  );
}

function parseUserAgent(ua) {
  if (!ua) return "Unknown Device";
  if (ua.includes("Firefox")) return "Firefox Browser";
  if (ua.includes("Chrome")) return "Chrome Browser";
  if (ua.includes("Safari")) return "Safari Browser";
  if (ua.includes("Edge")) return "Edge Browser";
  return "Web Browser";
}

export default function UserSettingsModal({ open, onClose }) {
  const { user } = useAuth();
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState(0);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [changingPassword, setChangingPassword] = useState(false);

  // Active Sessions Query
  const {
    data: sessionsList,
    isLoading: loadingSessions,
    isError: sessionsError,
    error: sessionsErrObj,
    refetch: refetchSessions,
  } = useQuery({
    queryKey: ["user-sessions"],
    queryFn: async () => {
      const res = await api.get("/sessions");
      return res.data?.data || [];
    },
    enabled: Boolean(open && activeTab === 1),
  });

  // Revoke Single Session Mutation
  const revokeSessionMutation = useMutation({
    mutationFn: async (sessionId) => {
      await api.delete(`/sessions/${sessionId}`);
    },
    onSuccess: () => {
      toast.success("Session revoked successfully");
      refetchSessions();
    },
    onError: (err) => {
      toast.error(err.message || "Failed to revoke session");
    },
  });

  // Revoke All Other Sessions Mutation
  const revokeOthersMutation = useMutation({
    mutationFn: async () => {
      await api.post("/sessions/revoke-others");
    },
    onSuccess: () => {
      toast.success("All other sessions have been revoked");
      refetchSessions();
    },
    onError: (err) => {
      toast.error(err.message || "Failed to revoke other sessions");
    },
  });

  // Change Password Form Submission
  const handleChangePassword = async (e) => {
    e.preventDefault();

    if (!currentPassword) {
      toast.error("Please enter your current password");
      return;
    }
    if (!newPassword || newPassword.length < 8) {
      toast.error("New password must be at least 8 characters long");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("Passwords do not match");
      return;
    }

    setChangingPassword(true);
    try {
      await api.put("/users/me/password", {
        current_password: currentPassword,
        new_password: newPassword,
        confirm_password: confirmPassword,
      });
      toast.success("Password updated successfully!");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      toast.error(err.message || "Failed to update password");
    } finally {
      setChangingPassword(false);
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
      <DialogTitle fontWeight={700}>Account Settings</DialogTitle>

      <DialogContent sx={{ pt: 0 }}>
        <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
          <Tabs
            value={activeTab}
            onChange={(e, val) => setActiveTab(val)}
            textColor="primary"
            indicatorColor="primary"
          >
            <Tab
              label="Password & Security"
              sx={{ fontWeight: 600, textTransform: "none" }}
            />
            <Tab
              label="Active Sessions"
              sx={{ fontWeight: 600, textTransform: "none" }}
            />
          </Tabs>
        </Box>

        {/* TAB 0: Password & Security */}
        <TabPanel value={activeTab} index={0}>
          <Box
            sx={{
              p: 2,
              mb: 3,
              backgroundColor: "#f8fafc",
              borderRadius: 2,
              border: "1px solid",
              borderColor: "divider",
            }}
          >
            <Typography
              variant="caption"
              color="text.secondary"
              display="block"
            >
              Signed in Email
            </Typography>
            <Typography
              variant="body2"
              fontWeight={600}
              gutterBottom
              color="text.primary"
            >
              {user?.email}
            </Typography>
            <Chip
              label={`Role: ${user?.role || "user"}`}
              size="small"
              sx={{
                fontWeight: 600,
                fontSize: "0.7rem",
                textTransform: "uppercase",
                mt: 0.5,
              }}
            />
          </Box>

          <Box component="form" onSubmit={handleChangePassword} noValidate>
            <Typography variant="subtitle2" fontWeight={600} gutterBottom>
              Change Password
            </Typography>

            <TextField
              margin="dense"
              required
              fullWidth
              name="currentPassword"
              label="Current Password"
              type="password"
              id="currentPassword"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              disabled={changingPassword}
            />

            <TextField
              margin="dense"
              required
              fullWidth
              name="newPassword"
              label="New Password"
              type="password"
              id="newPassword"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              disabled={changingPassword}
              helperText="Minimum 8 characters"
              sx={{ mt: 1.5 }}
            />

            <TextField
              margin="dense"
              required
              fullWidth
              name="confirmPassword"
              label="Confirm New Password"
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              disabled={changingPassword}
              sx={{ mt: 1.5 }}
            />

            <Button
              type="submit"
              variant="contained"
              disabled={changingPassword}
              sx={{ mt: 3, px: 3, py: 1 }}
            >
              {changingPassword ? (
                <CircularProgress size={20} color="inherit" />
              ) : (
                "Update Password"
              )}
            </Button>
          </Box>
        </TabPanel>

        {/* TAB 1: Active Sessions */}
        <TabPanel value={activeTab} index={1}>
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              mb: 2,
            }}
          >
            <Typography variant="body2" color="text.secondary">
              Manage your active login sessions across devices and browsers.
            </Typography>

            <Button
              variant="outlined"
              color="error"
              size="small"
              onClick={() => revokeOthersMutation.mutate()}
              disabled={revokeOthersMutation.isPending}
              sx={{ textTransform: "none", fontWeight: 600 }}
            >
              Revoke Others
            </Button>
          </Box>

          {sessionsError && (
            <Alert severity="error" sx={{ mb: 2, borderRadius: 2 }}>
              {sessionsErrObj?.message || "Failed to load sessions"}
            </Alert>
          )}

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
                  <TableCell sx={{ fontWeight: 600 }}>
                    Device / Browser
                  </TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>IP Address</TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>Logged In</TableCell>
                  <TableCell align="right" sx={{ fontWeight: 600 }}>
                    Actions
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {loadingSessions ? (
                  <TableRow>
                    <TableCell colSpan={4} align="center" sx={{ py: 3 }}>
                      <CircularProgress size={24} />
                    </TableCell>
                  </TableRow>
                ) : sessionsList?.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={4}
                      align="center"
                      sx={{ py: 3, color: "text.secondary" }}
                    >
                      No active sessions found.
                    </TableCell>
                  </TableRow>
                ) : (
                  sessionsList?.map((sess) => (
                    <TableRow key={sess.id}>
                      <TableCell sx={{ fontWeight: 600 }}>
                        <Box
                          sx={{ display: "flex", alignItems: "center", gap: 1 }}
                        >
                          <Typography variant="body2" fontWeight={600}>
                            {parseUserAgent(sess.user_agent)}
                          </Typography>
                          {sess.is_current && (
                            <Chip
                              label="Current"
                              size="small"
                              sx={{
                                height: 18,
                                fontSize: "0.65rem",
                                fontWeight: 600,
                                backgroundColor: "#dcfce7",
                                color: "#15803d",
                              }}
                            />
                          )}
                        </Box>
                      </TableCell>

                      <TableCell
                        sx={{ fontFamily: "monospace", fontSize: "0.8rem" }}
                      >
                        {sess.ip_address || "127.0.0.1"}
                      </TableCell>

                      <TableCell
                        sx={{ fontSize: "0.8rem", color: "text.secondary" }}
                      >
                        {sess.created_at
                          ? new Date(sess.created_at).toLocaleDateString()
                          : "-"}
                      </TableCell>

                      <TableCell align="right">
                        {!sess.is_current && (
                          <Button
                            size="small"
                            color="error"
                            onClick={() =>
                              revokeSessionMutation.mutate(sess.id)
                            }
                            disabled={revokeSessionMutation.isPending}
                            sx={{
                              textTransform: "none",
                              fontWeight: 600,
                              fontSize: "0.75rem",
                            }}
                          >
                            Revoke
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
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} variant="outlined" color="inherit">
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
}
