import { useParams } from "react-router-dom";

export default function PreviewConfigDetailPage() {
  const { configId } = useParams();
  return (
    <div className="p-6">
      <h1 className="text-lg font-semibold">Repository configuration</h1>
      <p className="text-sm text-muted-foreground">{configId}</p>
    </div>
  );
}
