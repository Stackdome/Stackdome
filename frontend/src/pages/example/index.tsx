import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function ExamplePage() {
  return (
    <div>
      <Card>
        <CardHeader>
          <CardTitle>Example Page</CardTitle>
          <CardDescription>This is an example page using the new layout</CardDescription>
        </CardHeader>
        <CardContent>
          <p>This page demonstrates the new layout with:</p>
          <ul className="list-disc pl-6 mt-2">
            <li>Collapsible sidebar with proper muted colors</li>
            <li>Breadcrumb navigation</li>
            <li>Collapsible icon with separator</li>
            <li>Content area with proper padding</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
