/*
Copyright (c) 2017, The University of Bristol, Senate House, Tyndall Avenue, Bristol, BS8 1TH, United Kingdom.
Copyright (c) 2021, COSIC-KU Leuven, Kasteelpark Arenberg 10, bus 2452, B-3001 Leuven-Heverlee, Belgium.

All rights reserved
*/

#include "Input_Output_Simple.h"
#include "Exceptions/Exceptions.h"

long Input_Output_Simple::open_channel(unsigned int channel)
{
  cout << "Opening channel " << channel << endl;
  return 0;
}

void Input_Output_Simple::close_channel(unsigned int channel)
{
  cout << "Closing channel " << channel << endl;
}

gfp Input_Output_Simple::private_input_gfp(unsigned int channel)
{
  cout << "Input channel " << channel << " : ";
  word x;
  cin >> x;
  gfp y;
  y.assign(x);
  return y;
}

void Input_Output_Simple::private_output_gfp(const gfp &output, unsigned int channel)
{
  cout << "Output channel " << channel << " : ";
  output.output(cout, true);
  cout << endl;
}

gfp Input_Output_Simple::public_input_gfp(unsigned int channel)
{
  cout << "Enter value on channel " << channel << " : ";
  word x;
  cin >> x;
  gfp y;
  y.assign(x);

  // Important to have this call in each version of public_input_gfp
  Update_Checker(y, channel);

  return y;
}

void Input_Output_Simple::public_output_gfp(const gfp &output, unsigned int channel)
{
  cout << "Output channel " << channel << " : ";
  output.output(cout, true);
  cout << endl;
}

long Input_Output_Simple::public_input_int(unsigned int channel)
{
  cout << "Enter value on channel " << channel << " : ";
  long x;
  cin >> x;

  // Important to have this call in each version of public_input_gfp
  Update_Checker(x, channel);

  return x;
}

void Input_Output_Simple::public_output_int(const long output, unsigned int channel)
{
  cout << "Output channel " << channel << " : " << output;
  cout << " = 0x" << hex << output << dec << endl;
}

void Input_Output_Simple::output_share(const Share &S, unsigned int channel)
{
  ofstream output_shares_file;
  if (counter2 == 0)
    {
      output_shares_file.open("Input/output_shares" + std::to_string(S.get_player()) + ".txt");
    } else {
      output_shares_file.open("Input/output_shares" + std::to_string(S.get_player()) + ".txt", ios::app);
    }

  S.output(output_shares_file, true);
  counter2++;
}

Share Input_Output_Simple::input_share(unsigned int channel, unsigned int whoimi)
{
  string line;
  ifstream myfile;
  string name;
  name = "Input/input_shares" + std::to_string(whoimi) + ".txt";

  if (counter1 == 0){
      int size_of_vector = 0;
      myfile.open(name);
      while (std::getline(myfile, line)) {
          size_of_vector++;
        }
      myfile.close();

      myfile.open(name);
      for (int i = 0; i < size_of_vector; i++) {
          Share ss;
          ss.input(myfile, true);
          shares_vector.push_back(ss);
        }
      myfile.close();
    }

  counter1++;

  return shares_vector[counter1 - 1];
}

void Input_Output_Simple::trigger(Schedule &schedule)
{
  printf("Restart requested: Enter a number to proceed\n");
  int i;
  cin >> i;

  // Load new schedule file program streams, using the original
  // program name
  //
  // Here you could define programatically what the new
  // programs you want to run are, by directly editing the
  // public variables in the schedule object.
  unsigned int nthreads= schedule.Load_Programs();
  if (schedule.max_n_threads() < nthreads)
    {
      throw Processor_Error("Restart requires more threads, cannot do this");
    }
}

void Input_Output_Simple::debug_output(const stringstream &ss)
{
  printf("%s", ss.str().c_str());
  fflush(stdout);
}

void Input_Output_Simple::crash(unsigned int PC, unsigned int thread_num)
{
  printf("Crashing in thread %d at PC value %d\n", thread_num, PC);
  throw crash_requested();
}
