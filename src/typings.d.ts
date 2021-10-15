// Typings reference file, you can add your own global typings here
// https://www.typescriptlang.org/docs/handbook/writing-declaration-files.html
type CounterType = {
  id: string;
  source?: string;
  idx?: number;
  show: boolean;
  label: string;
  type: string;
  tooltip: string;
};

type RInfo = {
  InstanceID: string;
  Version:    string;
  Commit:     string;
  Branch:     string;
  BuildStamp: string;
};
