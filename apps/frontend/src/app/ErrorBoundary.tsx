import React from "react";

type State = { failed: boolean };

export class ErrorBoundary extends React.Component<React.PropsWithChildren, State> {
  state: State = { failed: false };

  static getDerivedStateFromError(): State {
    return { failed: true };
  }

  render() {
    if (this.state.failed) {
      return <main className="runtime-state" role="alert"><h1>Something went wrong</h1><button className="primary-button" onClick={() => window.location.reload()}>Reload</button></main>;
    }
    return this.props.children;
  }
}
