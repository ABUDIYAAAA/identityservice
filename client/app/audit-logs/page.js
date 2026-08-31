"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
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
import Skeleton from "@mui/material/Skeleton";
import Alert from "@mui/material/Alert";
import Tooltip from "@mui/material/Tooltip";
import ProtectedRoute from "@/components/ProtectedRoute";
import DashboardLayout from "@/components/DashboardLayout";
import StateInspectModal from "@/components/StateInspectModal";
import { useAuth } from "@/hooks/useAuth";
import api from "@/lib/api";

function AuditLogsContent() {
  const { user: currentUser } = useAuth();
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [selectedEvent, setSelectedEvent] = useState(null);

  const apiPage = page + 1;
  const isAdmin = currentUser?.role === "admin";

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["audit-logs", apiPage, rowsPerPage],
    queryFn: async () => {
      const res = await api.get("/audit-logs", {
        params: { page: apiPage, limit: rowsPerPage },
      });
      return res.data?.data;
    },
    enabled: Boolean(isAdmin),
  });

  const auditList = data?.audit_logs || [];
  const totalCount = data?.total_count || 0;

  const handleChangePage = (event, newPage) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

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
          Access Restricted: Audit logs are accessible to system administrators
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
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
          Audit Logs & Security Traces
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Real-time security events, user administration, and service audit logs
        </Typography>
      </Box>

      {isError && (
        <Alert severity="error" sx={{ mb: 3, borderRadius: 2 }}>
          {error?.message || "Failed to load audit logs"}
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
          <Table sx={{ minWidth: 750 }}>
            <TableHead sx={{ backgroundColor: "#f8fafc" }}>
              <TableRow>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Action
                </TableCell>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Actor
                </TableCell>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Service Scope
                </TableCell>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Status
                </TableCell>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  IP Address
                </TableCell>
                <TableCell sx={{ fontWeight: 600, color: "text.secondary" }}>
                  Timestamp
                </TableCell>
                <TableCell
                  align="right"
                  sx={{ fontWeight: 600, color: "text.secondary" }}
                >
                  State
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, idx) => (
                  <TableRow key={idx}>
                    <TableCell>
                      <Skeleton variant="rounded" width={120} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" width={100} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" width={90} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="rounded" width={60} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" width={90} height={24} />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" width={120} height={24} />
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
              ) : auditList.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} align="center" sx={{ py: 6 }}>
                    <Typography variant="body2" color="text.secondary">
                      No audit events recorded yet.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                auditList.map((log) => (
                  <TableRow
                    key={log.id || Math.random()}
                    hover
                    sx={{ "&:last-child td, &:last-child th": { border: 0 } }}
                  >
                    <TableCell>
                      <Chip
                        label={log.action_type}
                        size="small"
                        sx={{
                          fontWeight: 600,
                          fontSize: "0.725rem",
                          backgroundColor: log.action_type.includes("FAIL")
                            ? "#fee2e2"
                            : "#f1f5f9",
                          color: log.action_type.includes("FAIL")
                            ? "#b91c1c"
                            : "#0f172a",
                        }}
                      />
                    </TableCell>

                    <TableCell>
                      <Typography
                        variant="caption"
                        color="text.secondary"
                        display="block"
                      >
                        {log.actor_type || "user"}
                      </Typography>
                      {log.actor_id ? (
                        <Tooltip title={log.actor_id} placement="top">
                          <Typography
                            variant="body2"
                            fontWeight={600}
                            sx={{
                              fontFamily: "monospace",
                              fontSize: "0.75rem",
                            }}
                          >
                            {log.actor_id.substring(0, 8)}...
                          </Typography>
                        </Tooltip>
                      ) : (
                        "-"
                      )}
                    </TableCell>

                    <TableCell>
                      {log.service_id ? (
                        <Chip
                          label={log.service_id.substring(0, 8)}
                          size="small"
                          sx={{
                            fontFamily: "monospace",
                            fontSize: "0.7rem",
                            backgroundColor: "#e2e8f0",
                          }}
                        />
                      ) : (
                        <Typography variant="caption" color="text.secondary">
                          -
                        </Typography>
                      )}
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

                    <TableCell>
                      <Typography
                        variant="body2"
                        sx={{
                          fontFamily: "monospace",
                          fontSize: "0.75rem",
                          color: "text.secondary",
                        }}
                      >
                        {log.ip_address || "-"}
                      </Typography>
                    </TableCell>

                    <TableCell>
                      <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{ fontSize: "0.8rem" }}
                      >
                        {log.created_at
                          ? new Date(log.created_at).toLocaleString()
                          : "-"}
                      </Typography>
                    </TableCell>

                    <TableCell align="right">
                      <Button
                        size="small"
                        variant="outlined"
                        color="inherit"
                        onClick={() => setSelectedEvent(log)}
                        sx={{
                          borderRadius: 2,
                          textTransform: "none",
                          fontWeight: 600,
                          fontSize: "0.75rem",
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

        <TablePagination
          rowsPerPageOptions={[10, 25, 50]}
          component="div"
          count={totalCount}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
          sx={{ borderTop: "1px solid", borderColor: "divider" }}
        />
      </Paper>

      {/* State Inspector Modal */}
      {selectedEvent && (
        <StateInspectModal
          open={Boolean(selectedEvent)}
          onClose={() => setSelectedEvent(null)}
          event={selectedEvent}
        />
      )}
    </Box>
  );
}

export default function AuditLogsPage() {
  return (
    <ProtectedRoute>
      <DashboardLayout>
        <AuditLogsContent />
      </DashboardLayout>
    </ProtectedRoute>
  );
}
