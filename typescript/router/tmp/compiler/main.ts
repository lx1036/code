
if (require.main === module) {
  const args = process.argv.slice(2);
  
  console.log(process.argv, args);
  process.exitCode = 0;
}