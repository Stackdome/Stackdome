import { useParams } from "react-router-dom";
export default function TeamDetailPage() {
  const { teamName } = useParams();
  return <div data-testid="team-detail-page">Team: {teamName}</div>;
}
