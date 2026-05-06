import { useState, useEffect } from "react";
import { Plus, Trash2, Loader2, Eye, EyeOff } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { FieldShell } from "@/components/branded";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { SecretFormSchema } from "../schemas/secret-schema";
import type { Secret, SecretType, SecretData } from "../types";
import { ZodError } from "zod";

interface SecretFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (secret: Omit<Secret, "id" | "organisation_id" | "created_at" | "updated_at">) => void;
  isLoading: boolean;
  error: string | null;
  editingSecret?: Secret | null;
}

const SECRET_TYPES: { value: SecretType; label: string }[] = [
  { value: "Generic", label: "Generic" },
  { value: "DockerRegistry", label: "Docker Registry" },
  { value: "GitCredentials", label: "Git Credentials" },
];

export function SecretFormDialog({
  open,
  onOpenChange,
  onSubmit,
  isLoading,
  error,
  editingSecret,
}: SecretFormDialogProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [type, setType] = useState<SecretType>("Generic");
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});
  const [showPassword, setShowPassword] = useState<Record<string, boolean>>({});

  // Type-specific fields
  const [registry, setRegistry] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [token, setToken] = useState("");
  const [sshPrivateKey, setSshPrivateKey] = useState("");
  const [genericData, setGenericData] = useState<SecretData[]>([{ key: "", value: "" }]);

  const isEditing = !!editingSecret;

  // Reset form when dialog opens/closes or when editing secret changes
  useEffect(() => {
    if (open) {
      if (editingSecret) {
        setName(editingSecret.name);
        setDescription(editingSecret.description || "");
        setType(editingSecret.type);

        // Populate type-specific fields based on existing data
        const data = editingSecret.data || [];
        data.forEach(item => {
          switch (item.key) {
            case "registry":
              setRegistry(item.value);
              break;
            case "username":
              setUsername(item.value);
              break;
            case "password":
              setPassword(item.value);
              break;
            case "token":
              setToken(item.value);
              break;
            case "ssh_private_key":
              setSshPrivateKey(item.value);
              break;
          }
        });

        // For generic type, populate the generic data array
        if (editingSecret.type === "Generic") {
          setGenericData(data.length > 0 ? data : [{ key: "", value: "" }]);
        }
      } else {
        // Reset all fields
        setName("");
        setDescription("");
        setType("Generic");
        setRegistry("");
        setUsername("");
        setPassword("");
        setToken("");
        setSshPrivateKey("");
        setGenericData([{ key: "", value: "" }]);
      }
      setFormErrors({});
      setShowPassword({});
    }
  }, [open, editingSecret]);

  const validateForm = (): boolean => {
    try {
      // Build the form data object based on type
      let formData: Record<string, unknown> = {
        name: name.trim(),
        description: description.trim() || undefined,
        type,
      };

      // Add type-specific fields
      switch (type) {
        case "DockerRegistry":
          formData = {
            ...formData,
            registry: registry.trim(),
            username: username.trim(),
            password: password.trim(),
          };
          break;
        case "GitCredentials":
          formData = {
            ...formData,
            username: username.trim() || undefined,
            password: password.trim() || undefined,
            token: token.trim() || undefined,
          };
          break;
        case "UsernamePassword":
          formData = {
            ...formData,
            username: username.trim(),
            password: password.trim(),
          };
          break;
        case "Token":
          formData = {
            ...formData,
            token: token.trim(),
          };
          break;
        case "SSHKey":
          formData = {
            ...formData,
            sshPrivateKey: sshPrivateKey.trim(),
          };
          break;
        case "Generic":
        default:
          formData = {
            ...formData,
            data: genericData.filter(item => item.key.trim() && item.value.trim()),
          };
          break;
      }

      // Validate with Zod
      SecretFormSchema.parse(formData);
      setFormErrors({});
      return true;
    } catch (error) {
      if (error instanceof ZodError) {
        const errors: Record<string, string> = {};

        error.issues.forEach((issue) => {
          const path = issue.path.join('.');

          // Handle array paths for generic data validation
          if (issue.path.length === 2 && issue.path[0] === 'data') {
            const index = issue.path[1];
            if (typeof index === 'number') {
              errors[`key-${index}`] = issue.message;
            }
          } else if (issue.path.length === 3 && issue.path[0] === 'data') {
            const index = issue.path[1];
            const field = issue.path[2];
            if (typeof index === 'number') {
              errors[`${field}-${index}`] = issue.message;
            }
          } else {
            errors[path || 'form'] = issue.message;
          }
        });

        setFormErrors(errors);
      }
      return false;
    }
  };

  const handleSubmit = () => {
    if (!validateForm()) return;

    let secretData: SecretData[] = [];

    // Build data array based on type
    switch (type) {
      case "DockerRegistry":
        secretData = [
          { key: "registry", value: registry.trim() },
          { key: "username", value: username.trim() },
          { key: "password", value: password.trim() },
        ];
        break;
      case "GitCredentials":
        if (username.trim() && password.trim()) {
          secretData = [
            { key: "username", value: username.trim() },
            { key: "password", value: password.trim() },
          ];
        } else if (token.trim()) {
          secretData = [{ key: "token", value: token.trim() }];
        }
        break;
      case "UsernamePassword":
        secretData = [
          { key: "username", value: username.trim() },
          { key: "password", value: password.trim() },
        ];
        break;
      case "Token":
        secretData = [{ key: "token", value: token.trim() }];
        break;
      case "SSHKey":
        secretData = [{ key: "ssh_private_key", value: sshPrivateKey.trim() }];
        break;
      case "Generic":
        secretData = genericData.filter(item => item.key.trim() && item.value.trim());
        break;
    }

    const secret = {
      name: name.trim(),
      description: description.trim() || undefined,
      type,
      data: secretData,
    };

    onSubmit(secret);
  };

  const addGenericDataPair = () => {
    setGenericData([...genericData, { key: "", value: "" }]);
  };

  const removeGenericDataPair = (index: number) => {
    if (genericData.length > 1) {
      setGenericData(genericData.filter((_, i) => i !== index));
    }
  };

  const updateGenericDataPair = (index: number, field: keyof SecretData, value: string) => {
    const updated = [...genericData];
    updated[index] = { ...updated[index], [field]: value };
    setGenericData(updated);

    // Clear field-specific errors when user starts typing
    const errorKey = `${field}-${index}`;
    if (formErrors[errorKey]) {
      setFormErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[errorKey];
        return newErrors;
      });
    }
  };

  const togglePasswordVisibility = (field: string) => {
    setShowPassword(prev => ({ ...prev, [field]: !prev[field] }));
  };

  const renderTypeSpecificFields = () => {
    switch (type) {
      case "DockerRegistry":
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="registry">Registry URL *</Label>
              <Input
                id="registry"
                value={registry}
                onChange={(e) => {
                  setRegistry(e.target.value);
                  if (formErrors.registry) {
                    setFormErrors(prev => ({ ...prev, registry: "" }));
                  }
                }}
                placeholder="e.g., docker.io, gcr.io, your-registry.com"
                className={formErrors.registry ? "border-danger" : ""}
              />
              {formErrors.registry && (
                <p className="text-sm text-danger">{formErrors.registry}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="username">Username *</Label>
              <Input
                id="username"
                value={username}
                onChange={(e) => {
                  setUsername(e.target.value);
                  if (formErrors.username) {
                    setFormErrors(prev => ({ ...prev, username: "" }));
                  }
                }}
                placeholder="Enter username"
                className={formErrors.username ? "border-danger" : ""}
              />
              {formErrors.username && (
                <p className="text-sm text-danger">{formErrors.username}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password *</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword.password ? "text" : "password"}
                  value={password}
                  onChange={(e) => {
                    setPassword(e.target.value);
                    if (formErrors.password) {
                      setFormErrors(prev => ({ ...prev, password: "" }));
                    }
                  }}
                  placeholder="Enter password"
                  className={formErrors.password ? "border-danger pr-10" : "pr-10"}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-0 top-0 h-full px-3"
                  onClick={() => togglePasswordVisibility("password")}
                >
                  {showPassword.password ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
              {formErrors.password && (
                <p className="text-sm text-danger">{formErrors.password}</p>
              )}
            </div>
          </div>
        );

      case "GitCredentials":
        return (
          <div className="space-y-4">
            {formErrors.credentials && (
              <div className="text-sm text-danger bg-danger-bg p-3 rounded-md">
                {formErrors.credentials}
              </div>
            )}
            <div className="space-y-4">
              <h4 className="text-sm font-medium">Option 1: Username & Password</h4>
              <div className="space-y-2">
                <Label htmlFor="username">Username</Label>
                <Input
                  id="username"
                  value={username}
                  onChange={(e) => {
                    setUsername(e.target.value);
                    if (formErrors.credentials) {
                      setFormErrors(prev => ({ ...prev, credentials: "" }));
                    }
                  }}
                  placeholder="Enter username"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <div className="relative">
                  <Input
                    id="password"
                    type={showPassword.password ? "text" : "password"}
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      if (formErrors.credentials) {
                        setFormErrors(prev => ({ ...prev, credentials: "" }));
                      }
                    }}
                    placeholder="Enter password"
                    className="pr-10"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute right-0 top-0 h-full px-3"
                    onClick={() => togglePasswordVisibility("password")}
                  >
                    {showPassword.password ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </Button>
                </div>
              </div>
              <div className="text-center text-sm text-muted-foreground">OR</div>
              <h4 className="text-sm font-medium">Option 2: Personal Access Token</h4>
              <div className="space-y-2">
                <Label htmlFor="token">Token</Label>
                <div className="relative">
                  <Input
                    id="token"
                    type={showPassword.token ? "text" : "password"}
                    value={token}
                    onChange={(e) => {
                      setToken(e.target.value);
                      if (formErrors.credentials) {
                        setFormErrors(prev => ({ ...prev, credentials: "" }));
                      }
                    }}
                    placeholder="Enter personal access token"
                    className="pr-10"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute right-0 top-0 h-full px-3"
                    onClick={() => togglePasswordVisibility("token")}
                  >
                    {showPassword.token ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </Button>
                </div>
              </div>
            </div>
          </div>
        );

      case "UsernamePassword":
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">Username *</Label>
              <Input
                id="username"
                value={username}
                onChange={(e) => {
                  setUsername(e.target.value);
                  if (formErrors.username) {
                    setFormErrors(prev => ({ ...prev, username: "" }));
                  }
                }}
                placeholder="Enter username"
                className={formErrors.username ? "border-danger" : ""}
              />
              {formErrors.username && (
                <p className="text-sm text-danger">{formErrors.username}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password *</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword.password ? "text" : "password"}
                  value={password}
                  onChange={(e) => {
                    setPassword(e.target.value);
                    if (formErrors.password) {
                      setFormErrors(prev => ({ ...prev, password: "" }));
                    }
                  }}
                  placeholder="Enter password"
                  className={formErrors.password ? "border-danger pr-10" : "pr-10"}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-0 top-0 h-full px-3"
                  onClick={() => togglePasswordVisibility("password")}
                >
                  {showPassword.password ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
              {formErrors.password && (
                <p className="text-sm text-danger">{formErrors.password}</p>
              )}
            </div>
          </div>
        );

      case "Token":
        return (
          <div className="space-y-2">
            <Label htmlFor="token">Token *</Label>
            <div className="relative">
              <Input
                id="token"
                type={showPassword.token ? "text" : "password"}
                value={token}
                onChange={(e) => {
                  setToken(e.target.value);
                  if (formErrors.token) {
                    setFormErrors(prev => ({ ...prev, token: "" }));
                  }
                }}
                placeholder="Enter token (min 8 characters)"
                className={formErrors.token ? "border-danger pr-10" : "pr-10"}
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="absolute right-0 top-0 h-full px-3"
                onClick={() => togglePasswordVisibility("token")}
              >
                {showPassword.token ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </Button>
            </div>
            {formErrors.token && (
              <p className="text-sm text-danger">{formErrors.token}</p>
            )}
          </div>
        );

      case "SSHKey":
        return (
          <div className="space-y-2">
            <Label htmlFor="sshPrivateKey">SSH Private Key *</Label>
            <Textarea
              id="sshPrivateKey"
              value={sshPrivateKey}
              onChange={(e) => {
                setSshPrivateKey(e.target.value);
                if (formErrors.sshPrivateKey) {
                  setFormErrors(prev => ({ ...prev, sshPrivateKey: "" }));
                }
              }}
              placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
              className={formErrors.sshPrivateKey ? "border-danger [field-sizing:fixed]" : "[field-sizing:fixed]"}
              rows={6}
            />
            {formErrors.sshPrivateKey && (
              <p className="text-sm text-danger">{formErrors.sshPrivateKey}</p>
            )}
          </div>
        );

      case "Generic":
      default:
        return (
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <Label className="text-[13px] font-medium text-foreground">
                <span>
                  Secret Data
                  <span className="ml-0.5 text-[15px] font-semibold text-brand/80 leading-none" aria-hidden>*</span>
                </span>
              </Label>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addGenericDataPair}
              >
                <Plus className="h-4 w-4" />
                Add Key-Value Pair
              </Button>
            </div>

            <div className="max-h-[200px] overflow-y-auto space-y-3 pr-2">
              {genericData.map((item, index) => (
                <div key={index} className="flex gap-2 items-start">
                  <div className="flex-1 space-y-1">
                    <Input
                      value={item.key}
                      onChange={(e) => updateGenericDataPair(index, "key", e.target.value)}
                      placeholder="Key"
                      className={formErrors[`key-${index}`] ? "border-danger" : ""}
                    />
                    {formErrors[`key-${index}`] && (
                      <p className="text-xs text-danger">{formErrors[`key-${index}`]}</p>
                    )}
                  </div>
                  <div className="flex-1 space-y-1">
                    <div className="relative">
                      <Input
                        value={item.value}
                        onChange={(e) => updateGenericDataPair(index, "value", e.target.value)}
                        placeholder="Value"
                        type={showPassword[`value-${index}`] ? "text" : "password"}
                        className={formErrors[`value-${index}`] ? "border-danger pr-10" : "pr-10"}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="absolute right-0 top-0 h-full px-3"
                        onClick={() => togglePasswordVisibility(`value-${index}`)}
                      >
                        {showPassword[`value-${index}`] ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </Button>
                    </div>
                    {formErrors[`value-${index}`] && (
                      <p className="text-xs text-danger">{formErrors[`value-${index}`]}</p>
                    )}
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeGenericDataPair(index)}
                    disabled={genericData.length === 1}
                    className="text-danger hover:text-danger hover:bg-danger-bg"
                    title="Remove key-value pair"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
            {formErrors.data && (
              <p className="text-sm text-danger">{formErrors.data}</p>
            )}
          </div>
        );
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[550px] max-w-[90vw] max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader className="flex-shrink-0">
          <DialogTitle>
            {isEditing ? "Edit Secret" : "Create New Secret"}
          </DialogTitle>
          <DialogDescription>
            {isEditing
              ? "Update the secret's metadata or rotate its values."
              : "Securely store API keys, passwords, or certificates for your stacks."}
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto">
          <div className="space-y-5 py-4">
            {error && (
              <div className="text-sm text-danger bg-danger-bg p-3 rounded-md">
                {error}
              </div>
            )}

            <FieldShell
              label="Name"
              htmlFor="name"
              required
              error={formErrors.name}
            >
              <Input
                id="name"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  if (formErrors.name) {
                    setFormErrors(prev => ({ ...prev, name: "" }));
                  }
                }}
                placeholder="Enter secret name"
                className={formErrors.name ? "border-danger" : ""}
              />
            </FieldShell>

            <FieldShell label="Description" htmlFor="description">
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Enter secret description (optional)"
                rows={2}
                className="resize-none [field-sizing:fixed]"
              />
            </FieldShell>

            <FieldShell label="Type" htmlFor="type" required>
              <Select value={type} onValueChange={(value: SecretType) => setType(value)}>
                <SelectTrigger id="type" className="w-full">
                  <SelectValue placeholder="Select secret type" />
                </SelectTrigger>
                <SelectContent>
                  {SECRET_TYPES.map((secretType) => (
                    <SelectItem key={secretType.value} value={secretType.value}>
                      {secretType.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FieldShell>

            {renderTypeSpecificFields()}
          </div>
        </div>

        <DialogFooter className="flex-shrink-0">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isLoading}>
            {isLoading && <Loader2 className="h-4 w-4 animate-spin" />}
            {isEditing ? "Update Secret" : "Create Secret"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
