import { useState } from 'react';
import { X, Code, Github, ChevronRight, ChevronsUpDown } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useNavigate } from 'react-router-dom';
import { getSampleStackYaml } from '@/lib/yaml-parser';
import { useStacks } from "@/pages/stacks/contexts/stack-context";

const StackCreatePage = () => {
  const navigate = useNavigate();
  const { addStack } = useStacks();
  const [currentStep, setCurrentStep] = useState(1);
  const [stackConfig, setStackConfig] = useState({
    name: '',
    description: '',
    region: 'US East (N. Virginia)',
    template: '',
    repositoryUrl: '',
    yamlConfig: getSampleStackYaml()
  });

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setStackConfig(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const selectTemplate = (template: string) => {
    setStackConfig(prev => ({
      ...prev,
      template
    }));
  };

  const goToNextStep = () => {
    setCurrentStep(prev => Math.min(prev + 1, 5));
  };

  const goToPrevStep = () => {
    setCurrentStep(prev => Math.max(prev - 1, 1));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Add the stack to our context
    addStack({
      name: stackConfig.name,
      description: stackConfig.description,
      template: stackConfig.template,
    });
    
    // Redirect to stacks page after creation
    navigate('/stacks');
  };

  return (
    <div className="fixed inset-0 bg-white z-40">
      <div className="border-b border-gray-200 p-4 flex justify-between items-center">
        <div className="flex items-center">
          <button 
            className="mr-4 text-gray-500"
            onClick={() => navigate('/stacks')}
          >
            <X size={20} />
          </button>
          <h1 className="text-lg font-medium">Create New Stack</h1>
        </div>
        <Button 
          className="bg-black text-white" 
          onClick={handleSubmit}
          disabled={currentStep < 5}
        >
          Deploy Stack
        </Button>
      </div>
      
      <div className="flex h-[calc(100vh-64px)]">
        {/* Sidebar */}
        <div className="w-64 border-r border-gray-200 bg-gray-50 p-4">
          <div className="space-y-1">
            {['Basic Info', 'Source Code', 'Stack Configuration', 'Environment', 'Review & Deploy'].map((step, index) => (
              <div 
                key={index}
                className={`flex items-center px-3 py-2 rounded-md text-sm ${
                  currentStep === index + 1 
                    ? 'bg-gray-200 font-medium' 
                    : 'text-gray-700'
                }`}
                onClick={() => setCurrentStep(index + 1)}
                style={{ cursor: 'pointer' }}
              >
                <span className={`mr-2 w-5 h-5 rounded-full flex items-center justify-center text-xs ${
                  currentStep === index + 1 
                    ? 'bg-black text-white' 
                    : 'bg-gray-200 text-gray-500'
                }`}>
                  {index + 1}
                </span>
                <span>{step}</span>
              </div>
            ))}
          </div>
        </div>
        
        {/* Main Content */}
        <div className="flex-grow p-6 overflow-y-auto">
          <div className="max-w-3xl mx-auto">
            {currentStep === 1 && (
              <>
                <h2 className="text-xl font-medium mb-6">Basic Information</h2>
                
                <div className="space-y-6">
                  <div>
                    <Label htmlFor="name">Stack Name</Label>
                    <Input
                      id="name"
                      name="name"
                      value={stackConfig.name}
                      onChange={handleInputChange}
                      placeholder="my-awesome-app"
                      className="mt-1"
                    />
                    <p className="mt-1 text-sm text-gray-500">This will be used as your stack identifier</p>
                  </div>
                  
                  <div>
                    <Label htmlFor="description">Description</Label>
                    <Textarea
                      id="description"
                      name="description"
                      value={stackConfig.description}
                      onChange={handleInputChange}
                      rows={3}
                      placeholder="Describe your stack (optional)"
                      className="mt-1"
                    />
                  </div>
                  
                  <div>
                    <Label htmlFor="region">Region</Label>
                    <div className="relative mt-1">
                      <select 
                        id="region"
                        name="region"
                        value={stackConfig.region}
                        onChange={handleInputChange}
                        className="w-full appearance-none px-3 py-2 border border-gray-300 rounded-md pr-10"
                      >
                        <option>US East (N. Virginia)</option>
                        <option>US West (Oregon)</option>
                        <option>EU West (Ireland)</option>
                        <option>Asia Pacific (Mumbai)</option>
                      </select>
                      <div className="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none">
                        <ChevronsUpDown size={16} className="text-gray-400" />
                      </div>
                    </div>
                  </div>
                  
                  <div>
                    <Label>Stack Template</Label>
                    <div className="grid grid-cols-3 gap-4 mt-2">
                      {['Next.js', 'Express', 'Django', 'Rails', 'Flutter'].map((template, idx) => (
                        <div 
                          key={idx}
                          className={`border rounded-md p-4 cursor-pointer hover:bg-gray-50 ${
                            stackConfig.template === template ? 'border-black' : ''
                          }`}
                          onClick={() => selectTemplate(template)}
                        >
                          <div className="flex items-center mb-2">
                            <Code size={18} className={`mr-2 ${
                              idx === 0 ? 'text-blue-500' :
                              idx === 1 ? 'text-purple-500' :
                              idx === 2 ? 'text-green-500' :
                              idx === 3 ? 'text-red-500' :
                              'text-blue-700'
                            }`} />
                            <span className="font-medium">{template}</span>
                          </div>
                          <p className="text-xs text-gray-500">
                            {idx === 0 ? 'React framework with SSR' :
                             idx === 1 ? 'Node.js web application' :
                             idx === 2 ? 'Python web framework' :
                             idx === 3 ? 'Ruby web framework' :
                             'Mobile app framework'
                            }
                          </p>
                        </div>
                      ))}
                      <div 
                        className="border border-dashed rounded-md p-4 cursor-pointer hover:bg-gray-50 flex items-center justify-center"
                        onClick={() => selectTemplate('Custom')}
                      >
                        <span className="text-gray-500">Custom Template</span>
                      </div>
                    </div>
                  </div>
                </div>
              </>
            )}
            
            {currentStep === 2 && (
              <>
                <h2 className="text-xl font-medium mb-6">Source Code</h2>
                
                <div className="space-y-6">
                  <div>
                    <Label htmlFor="repository">Source Repository</Label>
                    <div className="flex mt-1">
                      <Button variant="outline" className="flex items-center rounded-r-none">
                        <Github size={18} className="mr-2" />
                        <span>GitHub</span>
                      </Button>
                      <Input
                        id="repositoryUrl"
                        name="repositoryUrl"
                        value={stackConfig.repositoryUrl}
                        onChange={handleInputChange}
                        className="flex-grow border-l-0 rounded-l-none"
                        placeholder="username/repository"
                      />
                    </div>
                    <p className="mt-1 text-sm text-gray-500">Repository containing your application code</p>
                  </div>
                  
                  <div className="p-4 bg-gray-50 rounded-md border border-gray-200">
                    <h3 className="font-medium mb-2">Repository Connection</h3>
                    <p className="text-sm text-gray-600 mb-4">
                      To use a private repository, you'll need to set up authentication.
                    </p>
                    <Button variant="outline" size="sm">
                      Configure Git Credentials
                    </Button>
                  </div>
                  
                  <div className="p-4 bg-gray-50 rounded-md border border-gray-200">
                    <h3 className="font-medium mb-2">Branch</h3>
                    <div className="relative">
                      <select className="w-full appearance-none px-3 py-2 border border-gray-300 rounded-md pr-10">
                        <option>main</option>
                        <option>development</option>
                        <option>production</option>
                      </select>
                      <div className="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none">
                        <ChevronsUpDown size={16} className="text-gray-400" />
                      </div>
                    </div>
                  </div>
                </div>
              </>
            )}

            {currentStep === 3 && (
              <>
                <h2 className="text-xl font-medium mb-6">Stack Configuration</h2>
                
                <div className="space-y-6">
                  <div className="mb-6">
                    <Label>Stack Configuration File</Label>
                    <div className="bg-gray-50 p-4 rounded-md border border-gray-200 mt-2">
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-medium text-sm">stack-compose.yaml</span>
                        <Button variant="outline" size="sm">Edit</Button>
                      </div>
                      <pre className="bg-gray-800 text-green-400 p-3 rounded-md text-xs overflow-x-auto h-80 overflow-y-auto">
                        {stackConfig.yamlConfig}
                      </pre>
                    </div>
                  </div>
                  
                  <div className="p-4 bg-gray-50 rounded-md border border-gray-200">
                    <h3 className="font-medium mb-2">Advanced Settings</h3>
                    <div className="space-y-2">
                      <div className="flex items-center">
                        <input type="checkbox" id="autoRestart" className="mr-2" />
                        <Label htmlFor="autoRestart" className="text-sm">Enable auto-restart on code changes</Label>
                      </div>
                      <div className="flex items-center">
                        <input type="checkbox" id="enableLogs" className="mr-2" />
                        <Label htmlFor="enableLogs" className="text-sm">Enable detailed logging</Label>
                      </div>
                    </div>
                  </div>
                </div>
              </>
            )}

            {currentStep === 4 && (
              <>
                <h2 className="text-xl font-medium mb-6">Environment</h2>
                
                <div className="space-y-6">
                  <div className="p-4 bg-gray-50 rounded-md border border-gray-200">
                    <h3 className="font-medium mb-2">Environment Variables</h3>
                    <p className="text-sm text-gray-600 mb-4">
                      Add environment variables for your stack. These will be available to all services.
                    </p>
                    <div className="space-y-2">
                      <div className="flex gap-2">
                        <Input placeholder="KEY" className="w-1/3" />
                        <Input placeholder="VALUE" className="w-2/3" />
                      </div>
                      <div className="flex gap-2">
                        <Input placeholder="KEY" className="w-1/3" />
                        <Input placeholder="VALUE" className="w-2/3" />
                      </div>
                      <Button variant="outline" size="sm" className="mt-2">
                        Add Variable
                      </Button>
                    </div>
                  </div>
                  
                  <div className="p-4 bg-gray-50 rounded-md border border-gray-200">
                    <h3 className="font-medium mb-2">Secrets Management</h3>
                    <p className="text-sm text-gray-600 mb-4">
                      Manage sensitive information separately from your code.
                    </p>
                    <Button variant="outline" size="sm">
                      Configure Secrets
                    </Button>
                  </div>
                </div>
              </>
            )}

            {currentStep === 5 && (
              <>
                <h2 className="text-xl font-medium mb-6">Review & Deploy</h2>
                
                <div className="space-y-6">
                  <div className="p-4 bg-gray-50 rounded-md border border-gray-200">
                    <h3 className="font-medium mb-2">Stack Summary</h3>
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-gray-600">Name:</span>
                        <span className="font-medium">{stackConfig.name || 'Not specified'}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-600">Template:</span>
                        <span className="font-medium">{stackConfig.template || 'Not specified'}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-600">Region:</span>
                        <span className="font-medium">{stackConfig.region}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-600">Repository:</span>
                        <span className="font-medium">{stackConfig.repositoryUrl || 'Not specified'}</span>
                      </div>
                    </div>
                  </div>
                  
                  <div className="p-4 bg-gray-50 rounded-md border border-gray-200">
                    <h3 className="font-medium mb-2">Deployment Options</h3>
                    <div className="space-y-2">
                      <div className="flex items-center">
                        <input type="checkbox" id="autoScaling" className="mr-2" />
                        <Label htmlFor="autoScaling" className="text-sm">Enable auto-scaling</Label>
                      </div>
                      <div className="flex items-center">
                        <input type="checkbox" id="monitoring" className="mr-2" checked />
                        <Label htmlFor="monitoring" className="text-sm">Enable monitoring</Label>
                      </div>
                      <div className="flex items-center">
                        <input type="checkbox" id="backups" className="mr-2" />
                        <Label htmlFor="backups" className="text-sm">Enable automated backups</Label>
                      </div>
                    </div>
                  </div>
                  
                  <div className="p-4 bg-yellow-50 rounded-md border border-yellow-200">
                    <div className="flex items-start">
                      <span className="text-yellow-500 mr-2">⚠️</span>
                      <div>
                        <h3 className="font-medium mb-1">Important Notes</h3>
                        <p className="text-sm text-gray-600">
                          Deploying this stack may incur charges based on your usage. Make sure you have reviewed all configurations before deploying.
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </>
            )}
            
            <div className="mt-8 flex justify-between">
              {currentStep > 1 && (
                <Button 
                  variant="outline" 
                  onClick={goToPrevStep}
                >
                  Back
                </Button>
              )}
              
              {currentStep < 5 ? (
                <Button 
                  className="ml-auto flex items-center bg-black text-white"
                  onClick={goToNextStep}
                >
                  <span>Continue</span>
                  <ChevronRight size={16} className="ml-1" />
                </Button>
              ) : (
                <Button 
                  className="ml-auto bg-black text-white"
                  onClick={handleSubmit}
                >
                  Deploy Stack
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StackCreatePage;
