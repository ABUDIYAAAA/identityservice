"use client";

import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import Button from "@mui/material/Button";
import Chip from "@mui/material/Chip";
import Skeleton from "@mui/material/Skeleton";
import Alert from "@mui/material/Alert";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import ProtectedRoute from "@/components/ProtectedRoute";
import DashboardLayout from "@/components/DashboardLayout";
import InviteUserModal from "@/components/InviteUserModal";
import CreateServiceModal from "@/components/CreateServiceModal";
import SecretDisplayModal from "@/components/SecretDisplayModal";
import StateInspectModal from "@/components/StateInspectModal";
import { useAuth } from "@/hooks/useAuth";
import api from "@/lib/api";

function StatCard({ title, value, subtitle, color = "text.primary", loading }) {
  return (
    <Paper
      elevation={0}
      sx={{
        p: 3,
        borderRadius: 3,
        border: "1px solid",
        borderColor: "divider",
        backgroundColor: "#ffffff",
        height: "100%",
      }}
    >
      <Typography
        variant="caption"
        fontWeight={600}
        color="text.secondary"
        sx={{ textTransform: "uppercase" }}
      >
        {title}
      </Typography>
      <Typography variant="h3" fontWeight={700} color={color} sx={{ my: 1 }}>
        {loading ? <Skeleton width={60} /> : value}
      </Typography>
      {subtitle && (
        <Typography variant="caption" color="text.secondary">
          {subtitle}
        </Typography>
      )}
    </Paper>
  );
}

function HomePageContent() {
  const { user: currentUser } = useAuth();
  const queryClient = useQueryClient();

  const [inviteModalOpen, setInviteModalOpen] = useState(false);
  const [createServiceModalOpen, setCreateServiceModalOpen] = useState(false);
  const [initialSecretData, setInitialSecretData] = useState(null);
  const [selectedInspectEvent, setSelectedInspectEvent] = useState(null);

  const isAdmin = currentUser?.role === "admin";

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["dashboard-stats"],
    queryFn: async () => {
      const res = await api.get("/dashboard/stats");
      return res.data?.data;
    },
  });

  const handleServiceCreated = (createdData) => {
    queryClient.invalidateQueries(["dashboard-stats"]);
    if (createdData?.initial_secret) {
      setInitialSecretData({
        rawSecret: createdData.initial_secret,
        clientId: createdData.service?.client_id,
        serviceName: createdData.service?.name,
      });
    }
  };

  return (
    <Box>
      {/* Header Banner */}
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
          Executive System Overview
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Welcome back, <strong>{currentUser?.email}</strong> (
          {currentUser?.role})
        </Typography>
      </Box>

      {isError && (
        <Alert severity="error" sx={{ mb: 3, borderRadius: 2 }}>
          {error?.message || "Failed to load dashboard metrics"}
        </Alert>
      )}

      {/* Metrics Cards Grid */}
      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: {
            xs: "1fr",
            sm: "repeat(2, 1fr)",
            md: "repeat(4, 1fr)",
          },
          gap: 2.5,
          mb: 4,
        }}
      >
        <StatCard
          title="Active Users"
          value={data?.active_users_count ?? 0}
          subtitle="Enabled team members"
          loading={isLoading}
        />

        <StatCard
          title="Microservices"
          value={data?.services_count ?? 0}
          subtitle="Registered API services"
          loading={isLoading}
        />

        <StatCard
          title="Failed Logins (24h)"
          value={data?.failed_logins_24h ?? 0}
          subtitle="Security events last 24 hours"
          color={data?.failed_logins_24h > 0 ? "error.main" : "text.primary"}
          loading={isLoading}
        />

        <Paper
          elevation={0}
          sx={{
            p: 3,
            borderRadius: 3,
            border: "1px solid",
            borderColor: "divider",
            backgroundColor: "#ffffff",
            height: "100%",
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
          }}
        >
          <Box>
            <Typography
              variant="caption"
              fontWeight={600}
              color="text.secondary"
              sx={{ textTransform: "uppercase" }}
            >
              System Status
            </Typography>
            <Box
              sx={{ mt: 1.5, display: "flex", alignItems: "center", gap: 1 }}
            >
              <Chip
                label="Operational"
                size="small"
                sx={{
                  fontWeight: 700,
                  fontSize: "0.85rem",
                  backgroundColor: "#dcfce7",
                  color: "#15803d",
                  px: 1,
                }}
              />
            </Box>
          </Box>
          <Typography variant="caption" color="text.secondary" sx={{ mt: 1 }}>
            Redis Cache & DB Connected
          </Typography>
        </Paper>
      </Box>

      {/* Quick Action Shortcuts */}
      <Paper
        elevation={0}
        sx={{
          p: 3,
          borderRadius: 3,
          border: "1px solid",
          borderColor: "divider",
          backgroundColor: "#ffffff",
          mb: 4,
        }}
      >
        <Typography variant="subtitle1" fontWeight={700} gutterBottom>
          Quick Control Actions
        </Typography>

        <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1.5, mt: 2 }}>
          {isAdmin && (
            <Button
              variant="contained"
              disableElevation
              onClick={() => setInviteModalOpen(true)}
              sx={{ textTransform: "none", fontWeight: 600, px: 2.5 }}
            >
              + Invite Team User
            </Button>
          )}

          {isAdmin && (
            <Button
              variant="contained"
              disableElevation
              color="primary"
              onClick={() => setCreateServiceModalOpen(true)}
              sx={{ textTransform: "none", fontWeight: 600, px: 2.5 }}
            >
              + Register Service
            </Button>
          )}

          <Button
            component={Link}
            href="/services"
            variant="outlined"
            color="inherit"
            sx={{ textTransform: "none", fontWeight: 600, px: 2.5 }}
          >
            Manage Services
          </Button>

          {isAdmin && (
            <Button
              component={Link}
              href="/audit-logs"
              variant="outlined"
              color="inherit"
              sx={{ textTransform: "none", fontWeight: 600, px: 2.5 }}
            >
              View Audit Trail
            </Button>
          )}
        </Box>
      </Paper>

      {/* Recent Security & Audit Events Feed */}
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
        <Box
          sx={{
            p: 3,
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
          }}
        >
          <Box>
            <Typography variant="subtitle1" fontWeight={700}>
              Recent Security & Audit Activity
            </Typography>
            <Typography variant="caption" color="text.secondary">
              Latest events recorded in background audit pipeline
            </Typography>
          </Box>

          {isAdmin && (
            <Button
              component={Link}
              href="/audit-logs"
              size="small"
              color="inherit"
              sx={{ textTransform: "none", fontWeight: 600 }}
            >
              View All Logs &rarr;
            </Button>
          )}
        </Box>

        <TableContainer>
          <Table size="small">
            <TableHead sx={{ backgroundColor: "#f8fafc" }}>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>Action</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Actor</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Status</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Timestamp</TableCell>
                <TableCell align="right" sx={{ fontWeight: 600 }}>
                  Inspect
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {isLoading ? (
                Array.from({ length: 3 }).map((_, idx) => (
                  <TableRow key={idx}>
                    <TableCell>
                      <Skeleton width={120} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton width={90} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton width={60} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton width={100} height={24} />
                    </TableCell>
                    <TableCell align="right">
                      <Skeleton width={60} height={24} sx={{ ml: "auto" }} />
                    </TableCell>
                  </TableRow>
                ))
              ) : data?.recent_audit_logs?.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    align="center"
                    sx={{ py: 4, color: "text.secondary" }}
                  >
                    No recent audit activity.
                  </TableCell>
                </TableRow>
              ) : (
                data?.recent_audit_logs?.map((log) => (
                  <TableRow key={log.id || Math.random()} hover>
                    <TableCell>
                      <Chip
                        label={log.action_type}
                        size="small"
                        sx={{
                          fontWeight: 600,
                          fontSize: "0.725rem",
                          backgroundColor:
                            log.action_type.includes("FAIL") ||
                            log.action_type.includes("LOCKED")
                              ? "#fee2e2"
                              : "#f1f5f9",
                          color:
                            log.action_type.includes("FAIL") ||
                            log.action_type.includes("LOCKED")
                              ? "#b91c1c"
                              : "#0f172a",
                        }}
                      />
                    </TableCell>

                    <TableCell
                      sx={{ fontSize: "0.8rem", color: "text.secondary" }}
                    >
                      {log.actor_id
                        ? `${log.actor_type || "user"}:${log.actor_id.substring(0, 8)}`
                        : log.actor_type || "system"}
                    </TableCell>

                    <TableCell>
                      <Chip
                        label={log.status || "success"}
                        size="small"
                        sx={{
                          fontWeight: 600,
                          fontSize: "0.7rem",
                          backgroundColor:
                            log.status === "failure" ? "#fee2e2" : "#dcfce7",
                          color:
                            log.status === "failure" ? "#b91c1c" : "#15803d",
                        }}
                      />
                    </TableCell>

                    <TableCell
                      sx={{ fontSize: "0.8rem", color: "text.secondary" }}
                    >
                      {log.created_at
                        ? new Date(log.created_at).toLocaleTimeString([], {
                            hour: "2-digit",
                            minute: "2-digit",
                          })
                        : "-"}
                    </TableCell>

                    <TableCell align="right">
                      <Button
                        size="small"
                        variant="outlined"
                        color="inherit"
                        onClick={() => setSelectedInspectEvent(log)}
                        sx={{
                          borderRadius: 2,
                          textTransform: "none",
                          fontWeight: 600,
                          fontSize: "0.75rem",
                          py: 0.3,
                        }}
                      >
                        Inspect
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>

      {/* Dialog Modals */}
      <InviteUserModal
        open={inviteModalOpen}
        onClose={() => setInviteModalOpen(false)}
        onSuccess={() => queryClient.invalidateQueries(["dashboard-stats"])}
      />

      <CreateServiceModal
        open={createServiceModalOpen}
        onClose={() => setCreateServiceModalOpen(false)}
        onSuccess={handleServiceCreated}
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

      {selectedInspectEvent && (
        <StateInspectModal
          open={Boolean(selectedInspectEvent)}
          onClose={() => setSelectedInspectEvent(null)}
          event={selectedInspectEvent}
        />
      )}
    </Box>
  );
}

export default function HomePage() {
  return (
    <ProtectedRoute>
      <DashboardLayout>
        <HomePageContent />
      </DashboardLayout>
    </ProtectedRoute>
  );
}
