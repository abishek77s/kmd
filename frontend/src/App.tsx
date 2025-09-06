import Commands from "./components/Commands";
import File from "./components/File";
import Navbar from "./components/Navbar";

function App() {
    return (
        <>
            <div className="flex flex-col ">
                <Navbar />
                <div className="p-6">
                    <File />
                    <Commands />
                </div>
            </div>
        </>
    );
}

export default App;
