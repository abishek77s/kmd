import Code from "./components/ShareClone";
import Navbar from "./components/Navbar";

function App() {
  return (
    <>
      <div className="flex flex-col ">
        <Navbar />
        <div className="p-6">
          <Code />
        </div>
      </div>
    </>
  );
}

export default App;
