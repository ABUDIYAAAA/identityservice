"use client";

import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import ProtectedRoute from "@/components/ProtectedRoute";
import DashboardLayout from "@/components/DashboardLayout";
import { useAuth } from "@/hooks/useAuth";

function DashboardContent() {
  const { user } = useAuth();

  return (
    <Box>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
          Dashboard
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Welcome back, <strong>{user?.email}</strong>
        </Typography>
      </Box>

      {/* Quick Links Placeholder Container */}
      <Paper
        elevation={0}
        sx={{
          p: 6,
          borderRadius: 3,
          border: "2px dashed",
          borderColor: "divider",
          backgroundColor: "#ffffff",
          textAlign: "center",
          minHeight: 280,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <Typography variant="h6" fontWeight={600} color="text.secondary" gutterBottom>
          Quick Links & Actions
        </Typography>
        <Typography variant="body2" color="text.disabled" sx={{ maxWidth: 360 }}>
          This area will contain quick links and management widgets.
        </Typography>
      </Paper>
    </Box>
  );
}

export default function HomePage() {
  return (
    <ProtectedRoute>
      <DashboardLayout>
        <DashboardContent />
      </DashboardLayout>
    </ProtectedRoute>
  );
}
